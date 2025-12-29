package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dbunt1tled/fiber-go/internal/config"
	"github.com/dbunt1tled/fiber-go/internal/lib/aemail"
	"github.com/dbunt1tled/fiber-go/internal/lib/email"
	"github.com/dbunt1tled/fiber-go/internal/modules/auth"
	"github.com/dbunt1tled/fiber-go/internal/modules/user"
	"github.com/dbunt1tled/fiber-go/pkg/hasher"
	"github.com/dbunt1tled/fiber-go/pkg/http/er"
	"github.com/dbunt1tled/fiber-go/pkg/http/middlewares"
	"github.com/dbunt1tled/fiber-go/pkg/log"
	"github.com/dbunt1tled/fiber-go/pkg/queue"
	"github.com/dbunt1tled/fiber-go/pkg/validation"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/pprof"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type Application struct {
	engine *fiber.App
	cfg    *config.ServiceConfig

	MailService    *email.MailService
	AuthController *auth.Controller
	UserController *user.Controller

	AuthMiddleware *middlewares.AuthMiddleware
}

func NewApp(cfg *config.ServiceConfig) *Application {
	engine := engineSetup()
	validator, err := validation.Validator(cfg.DB.Pool())
	if err != nil {
		panic(err)
	}
	userService := user.NewUserService(cfg.DB.Pool())
	hashService, err := hasher.NewHasher(
		config.Get().Server.JWT.Algorithm,
		config.Get().Server.JWT.PublicKey,
		config.Get().Server.JWT.PrivateKey,
	)
	if err != nil {
		panic(err)
	}
	authService := auth.NewAuthService(hashService)
	mailService := email.NewMailService(cfg.Mailer, config.Get().Mailer.Address)
	mailServiceAsync := aemail.NewMailServiceAsync(cfg.Producer)
	return &Application{
		engine:         engine,
		cfg:            cfg,
		MailService:    mailService,
		AuthController: auth.NewController(authService, userService, mailServiceAsync, validator),
		UserController: user.NewUserController(userService, validator),
		AuthMiddleware: middlewares.NewAuthMiddleware(authService, userService),
	}
}

func engineSetup() *fiber.App {
	engine := fiber.New(fiber.Config{
		JSONEncoder:  sonic.ConfigFastest.Marshal,
		JSONDecoder:  sonic.ConfigFastest.Unmarshal,
		ReadTimeout:  config.Get().Server.HTTP.Timeout,
		WriteTimeout: config.Get().Server.HTTP.Timeout,
		ErrorHandler: er.APIErrorHandler,
	})

	// Middleware setup
	engine.Use(recover.New())
	engine.Use(middlewares.NewLog(func(c fiber.Ctx) bool {
		return false
	}))
	engine.Use(helmet.New())
	engine.Use(compress.New())
	if config.Get().Debug {
		engine.Use(pprof.New())
	}
	engine.Use(cors.New(cors.Config{
		AllowOrigins:  strings.Split(config.Get().Server.HTTP.CORS.AllowOrigins, ","),
		AllowMethods:  strings.Split(config.Get().Server.HTTP.CORS.AllowMethods, ","),
		AllowHeaders:  strings.Split(config.Get().Server.HTTP.CORS.AllowHeaders, ","),
		ExposeHeaders: strings.Split(config.Get().Server.HTTP.CORS.ExposeHeaders, ","),
	}))

	if config.Get().Static.URL != "" && config.Get().Static.Directory != "" {
		engine.Get(
			fmt.Sprintf("/%s/*", config.Get().Static.URL),
			static.New(config.Get().Static.Directory, static.Config{
				CacheDuration: 10 * time.Minute, //nolint:mnd // default
				MaxAge:        3600,             //nolint:mnd // default
			}),
		)
	}

	return engine
}

func (a *Application) Run(ctx context.Context) error {
	return a.serveWithGraceFullShutdown(
		ctx,
		config.Get().Server.HTTP.Host+":"+strconv.Itoa(config.Get().Server.HTTP.Port),
	)
}

func (a *Application) Engine() *fiber.App {
	return a.engine
}

func (a *Application) serveWithGraceFullShutdown(ctx context.Context, addr string) error {
	var wg sync.WaitGroup
	c, stop := signal.NotifyContext(
		ctx,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGABRT,
		syscall.SIGQUIT,
		os.Interrupt,
	)
	defer stop()
	emailHandler := queue.NewEmailHandler(a.MailService)
	mux := queue.NewQueueMux(emailHandler)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.cfg.Consumer.Run(mux); err != nil {
			log.Logger().Error("Consumer can't run.", err)
		}
	}()

	go func() {
		if err := a.engine.Listen(addr, fiber.ListenConfig{
			EnablePrefork: config.Get().Server.HTTP.Prefork,
		}); err != nil {
			panic(err)
		}
	}()
	<-c.Done()
	var cancel context.CancelFunc
	c, cancel = context.WithTimeout(context.Background(), 10*time.Second) //nolint:mnd // 10 seconds timeout
	defer cancel()

	log.Logger().Warn("Quit: shutting down ...")
	defer log.Logger().Warn("｡◕‿‿◕｡ Quit: shutdown completed")
	log.Logger().Warn("㋡ Quit: closing database connection")
	a.cfg.DB.Close()
	log.Logger().Warn("㋡ Quit: closing mailer connection")
	_ = a.cfg.Mailer.Close()
	log.Logger().Warn("㋡ Quit: closing producer")
	_ = a.cfg.Producer.Close()
	log.Logger().Warn("㋡ Quit: closing consumer")
	a.cfg.Consumer.Close()
	wg.Wait()
	log.Logger().Warn("㋡ Quit: closing logger")
	_ = log.Close()
	return a.engine.Shutdown()
}
