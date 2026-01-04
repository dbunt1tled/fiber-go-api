package config

import (
	"github.com/dbunt1tled/fiber-go-api/pkg/db"
	"github.com/dbunt1tled/fiber-go-api/pkg/mailer"
	"github.com/dbunt1tled/fiber-go-api/pkg/queue"
)

type ServiceConfig struct {
	DB       *db.DB
	Mailer   *mailer.Mailer
	Producer *queue.Producer
	Consumer *queue.Consumer
}

func NewServiceConfig(
	db *db.DB,
	mail *mailer.Mailer,
	producer *queue.Producer,
	consumer *queue.Consumer,
) *ServiceConfig {
	return &ServiceConfig{
		DB:       db,
		Mailer:   mail,
		Producer: producer,
		Consumer: consumer,
	}
}
