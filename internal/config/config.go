package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Listen struct {
		BindIP string `env:"BIND_IP" env-default:"0.0.0.0"`
		Port   string `env:"PORT" env-default:"8080"`
	}
	Server struct {
		IdleTimeout       int  `env:"IDLE_TIMEOUT" env-default:"100"`
		RateLimiterEnable bool `env:"RATE_LIMITER_ENABLE" env-default:"true"`

		//RateLimiterType - can be based on IP address (type "ip") or based on the number of requests (type "requests")
		RateLimiterType string `env:"RATE_LIMITER_TYPE" env-default:"requests"`

		RateLimiterMemoryStoreConfig struct {
			Rate      float64 `env:"RATE" env-default:"10"`
			Burst     int     `env:"BURST" env-default:"1"`
			ExpiresIn int     `env:"EXPIRES_IN" env-default:"180"`
		}
	}
	AppConfig struct {
		LogLevel    string `env:"LOGLEVEL" env-default:"debug"`
		CourierType struct {
			FootCourierType string `env:"FOOT_COURIER_TYPE" env-default:"FOOT"`
			BikeCourierType string `env:"BIKE_COURIER_TYPE" env-default:"BIKE"`
			AutoCourierType string `env:"AUTO_COURIER_TYPE" env-default:"AUTO"`
		}
		CourierCostFactor struct {
			FootCourierCostFactor int `env:"FOOT_COURIER_COST_FACTOR" env-default:"2"`
			BikeCourierCostFactor int `env:"BIKE_COURIER_COST_FACTOR" env-default:"3"`
			AutoCourierCostFactor int `env:"AUTO_COURIER_COST_FACTOR" env-default:"4"`
		}
		CourierRateFactor struct {
			FootCourierRateFactor int `env:"FOOT_COURIER_RATE_FACTOR" env-default:"3"`
			BikeCourierRateFactor int `env:"BIKE_COURIER_RATE_FACTOR" env-default:"2"`
			AutoCourierRateFactor int `env:"AUTO_COURIER_RATE_FACTOR" env-default:"1"`
		}
	}
	Database struct {
		PostgresDSN string `env:"POSTGRES_DSN" env-default:"postgresql://postgres:password@db/postgres"`
		SslMode     string `env:"SSL_MODE" env-default:"disable"`
	}
}

func New() (*Config, error) {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
