package client

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
	"yandex-lavka-service/internal/config"
	//_ "github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
)

type PostgresSql struct {
	Db     *sqlx.DB
	config *config.Config
}

func NewPostgresSql(config *config.Config) *PostgresSql {
	return &PostgresSql{config: config}
}

func (p *PostgresSql) Open() error {
	dbSourceName := p.config.Database.PostgresDSN + "?" + "sslmode=" + p.config.Database.SslMode
	db, err := sqlx.Connect("postgres", dbSourceName)
	if err != nil {
		return err
	}
	p.Db = db
	log.Info("Connection to db successfully")
	return nil
}

func (p *PostgresSql) Close() error {
	err := p.Db.Close()
	if err != nil {
		return err
	}
	log.Info("Close db successfully")
	return nil
}
