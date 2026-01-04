package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dbunt1tled/fiber-go-api/internal/lib/email"
	"github.com/dbunt1tled/fiber-go-api/pkg/log"
	"github.com/hibiken/asynq"
)

const (
	EmailQueue    = "emails"
	EmailSendTask = "email:send"
	MaxRetry      = 3
	TimeOut       = 30 * time.Second
)

type EmailPayload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

type EmailHandler struct {
	mailerService *email.MailService
}

func NewEmailHandler(mailerService *email.MailService) *EmailHandler {
	return &EmailHandler{
		mailerService: mailerService,
	}
}

func (e *EmailPayload) Data() ([]byte, error) {
	return sonic.ConfigFastest.Marshal(map[string]string{
		"to":      e.To,
		"subject": e.Subject,
		"body":    e.Body,
	})
}

func (p *Producer) SendEmail(payload *EmailPayload) error {
	data, err := payload.Data()
	if err != nil {
		return err
	}
	task := asynq.NewTask(EmailSendTask, data)

	info, err := p.client.Enqueue(task,
		asynq.MaxRetry(MaxRetry),
		asynq.Timeout(TimeOut),
		asynq.Queue(EmailQueue),
	)
	if err != nil {
		return err
	}
	logEnqueue(info)

	return nil
}

func (e *EmailHandler) SendEmailHandler(ctx context.Context, t *asynq.Task) error {
	var payload EmailPayload
	err := sonic.ConfigFastest.Unmarshal(t.Payload(), &payload)
	if err != nil {
		logError(t, "SEND EMAIL Error", err)
		return err
	}
	log.Logger().Info(fmt.Sprintf("SEND EMAIL receive task id=%s", t.ResultWriter().TaskID()))

	err = e.mailerService.SendEmail(ctx, payload.To, payload.Subject, payload.Body)
	if err != nil {
		logError(t, fmt.Sprintf("SEND EMAIL Error send to %s", payload.To), err)
		return err
	}
	logSuccess(t, fmt.Sprintf("SEND EMAIL Success send to %s", payload.To))

	return nil
}
