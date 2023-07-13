package yandexlavkaservice

import (
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/config"
	"yandex-lavka-service/internal/controller/httpapi"
)

type YandexLavkaService struct {
	config *config.Config
}

func New(config *config.Config) *YandexLavkaService {
	return &YandexLavkaService{config: config}
}

func (s *YandexLavkaService) Start() error {
	server, err := httpapi.NewServer(s.config)
	if err != nil {
		return err
	}
	go server.Start()
	log.Info("http server started")
	server.GracefulShutdown()
	return nil
}
