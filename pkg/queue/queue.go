package queue

import (
	"fmt"
	"log/slog"

	"github.com/dbunt1tled/fiber-go-api/pkg/log"
	"github.com/hibiken/asynq"
)

type Producer struct {
	client *asynq.Client
}

type Consumer struct {
	server *asynq.Server
}

func NewQueueProducer(redisAddr string) *Producer {
	return &Producer{
		client: asynq.NewClient(asynq.RedisClientOpt{
			Addr: redisAddr,
		}),
	}
}

func NewQueueConsumer(redisAddr string) *Consumer {
	return &Consumer{
		server: asynq.NewServer(
			asynq.RedisClientOpt{Addr: redisAddr},
			asynq.Config{
				Concurrency: 3, //nolint:mnd // for example
				Queues: map[string]int{
					EmailQueue: 2, //nolint:mnd // 2 + 1 = 3
					"default":  1,
				},
			}),
	}
}

func NewQueueMux(
	emailHandler *EmailHandler,
) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)
	mux.HandleFunc(EmailSendTask, emailHandler.SendEmailHandler)

	return mux
}

func (c *Consumer) Close() {
	c.server.Stop()
}

func (c *Consumer) Run(mux *asynq.ServeMux) error {
	return c.server.Run(mux)
}

func (p *Producer) Close() error {
	return p.client.Close()
}

func logError(t *asynq.Task, msg string, err error) {
	log.Logger().Error(
		msg,
		err,
		slog.String("id", t.ResultWriter().TaskID()),
		slog.String("payload", string(t.Payload())),
	)
}

func logSuccess(t *asynq.Task, msg string) {
	log.Logger().Info(
		msg,
		slog.String("payload", string(t.Payload())),
	)
}

func logEnqueue(t *asynq.TaskInfo) {
	log.Logger().Info(
		fmt.Sprintf("Enqueued task: id=%s queue=%s", t.ID, t.Queue),
		slog.String("id", t.ID),
		slog.String("action", t.Type),
		slog.String("queue", t.Queue),
		slog.String("payload", string(t.Payload)),
	)
}
