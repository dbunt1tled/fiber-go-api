package aemail

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/dbunt1tled/fiber-go/internal/config"
	"github.com/dbunt1tled/fiber-go/internal/lib/view"
	"github.com/dbunt1tled/fiber-go/internal/modules/user"
	"github.com/dbunt1tled/fiber-go/pkg/queue"
)

type MailServiceAsync struct {
	producer *queue.Producer
}

func NewMailServiceAsync(producer *queue.Producer) *MailServiceAsync {
	return &MailServiceAsync{
		producer: producer,
	}
}

func (m *MailServiceAsync) SendConfirmMail(user *user.User, token string) error {
	data := view.MakeTemplateData(map[string]any{
		"User":  *user,
		"Token": token,
	})
	templ, err := view.GetTemplate("auth/register.gohtml")
	if err != nil {
		return err
	}
	if templ == nil {
		return errors.New("template not found")
	}
	var html bytes.Buffer
	err = templ.Execute(&html, data)
	if err != nil {
		return err
	}

	payload := queue.EmailPayload{
		To:      user.Email,
		Subject: fmt.Sprintf("Welcome to %s", config.Get().Name),
		Body:    html.String(),
	}
	return m.producer.SendEmail(&payload)
}
