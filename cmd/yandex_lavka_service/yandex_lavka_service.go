package main

import (
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/app"
	"yandex-lavka-service/internal/config"
)

func main() {
	log.Info("init service")
	cfg, err := config.New()
	if err != nil {
		log.Fatal("error read config")
	}
	yandexLavka := yandexlavkaservice.New(cfg)
	err = yandexLavka.Start()
	if err != nil {
		log.Info(err)
		return
	}
}
