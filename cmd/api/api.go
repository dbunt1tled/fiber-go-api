package main

import (
	"context"

	"github.com/dbunt1tled/fiber-go-api/internal/app"
	"github.com/dbunt1tled/fiber-go-api/internal/app/routes"
	"github.com/dbunt1tled/fiber-go-api/internal/config"
	"github.com/dbunt1tled/fiber-go-api/pkg/db"
	"github.com/dbunt1tled/fiber-go-api/pkg/http/middlewares"
	"github.com/dbunt1tled/fiber-go-api/pkg/log"
	"github.com/dbunt1tled/fiber-go-api/pkg/mailer"
	"github.com/dbunt1tled/fiber-go-api/pkg/queue"
)

func main() {
	config.Load()
	log.Load(config.Get().Name, config.Get().Env, config.Get().Log.Level, config.Get().Log.File)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	database := db.New(ctx, config.Get().DB.Main.DSN)
	mail := mailer.NewMailer(
		config.Get().Mailer.Host,
		config.Get().Mailer.Port,
		config.Get().Mailer.Username,
		config.Get().Mailer.Password,
		config.Get().Mailer.Address,
	)
	producer := queue.NewQueueProducer(config.Get().Redis.Addr)
	consumer := queue.NewQueueConsumer(config.Get().Redis.Addr)
	cfg := config.NewServiceConfig(database, mail, producer, consumer)
	application := app.NewApp(cfg)
	routes.WebRoutes(application)
	routes.ApiRoutes(application)
	application.Engine().Use(middlewares.NotFound)
	err := application.Run(ctx)
	if err != nil {
		panic(err)
	}
}
