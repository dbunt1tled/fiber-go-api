package email

import (
	"context"

	"github.com/dbunt1tled/fiber-go-api/pkg/mailer"
	"github.com/wneessen/go-mail"
)

type MailService struct {
	mailer      *mailer.Mailer
	fromAddress string
}

func NewMailService(mailer *mailer.Mailer, fromAddress string) *MailService {
	return &MailService{
		mailer:      mailer,
		fromAddress: fromAddress,
	}
}

func (m *MailService) SendEmail(
	c context.Context,
	to string,
	subject string,
	body string,
) error {
	e := mail.NewMsg()

	err := e.From(m.fromAddress)
	if err != nil {
		return err
	}
	err = e.To(to)
	if err != nil {
		return err
	}

	e.Subject(subject)
	e.SetBodyString(mail.TypeTextHTML, body)

	if err = m.mailer.SendCtx(c, e); err != nil {
		return err
	}

	return nil
}
