package httpapi

import (
	"context"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"
	"net/http"
	"os"
	"os/signal"
	"time"
	"yandex-lavka-service/internal/config"
	"yandex-lavka-service/internal/controller/httpapi/handlers"
	"yandex-lavka-service/internal/domain/service"
	"yandex-lavka-service/internal/store/postgressql/adapters"
	"yandex-lavka-service/internal/store/postgressql/client"
)

type Server struct {
	echoFramework *echo.Echo
	courier       *handlers.Courier
	order         *handlers.Order
	clientDb      *client.PostgresSql
	config        *config.Config
	validate      *validator.Validate
}

func NewServer(config *config.Config) (*Server, error) {
	e := echo.New()

	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: time.Duration(config.Server.IdleTimeout) * time.Second,
	}))
	e.Use(middleware.Logger())

	var validate = validator.New()
	err := validate.RegisterValidation("time_interval_format", handlers.ValidateTimeInterval)
	if err != nil {
		return nil, err
	}

	clientDb := client.NewPostgresSql(config)
	err = clientDb.Open()
	if err != nil {
		return nil, err
	}

	courierRepository := adapters.NewCourierRepository(clientDb)
	courierService := service.NewCourier(courierRepository, config)
	courier := handlers.NewCourier(courierService, validate)

	orderRepository := adapters.NewOrderRepository(clientDb)
	orderService := service.NewOrder(orderRepository)
	order := handlers.NewOrder(orderService, validate)

	return &Server{echoFramework: e, courier: courier, order: order, clientDb: clientDb, config: config, validate: validate}, nil
}

func (s *Server) registerRouters() {
	if s.config.Server.RateLimiterEnable {
		s.echoFramework.POST("/couriers", s.courier.AddCouriers, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.GET("/couriers/:id", s.courier.GetCourier, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.GET("/couriers", s.courier.GetCouriers, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.GET("/couriers/meta-info/:id", s.courier.GetCourierMetaInfo, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))

		s.echoFramework.POST("/orders", s.order.AddOrders, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.POST("/orders/complete", s.order.SetCompleteOrders, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.GET("/orders/:id", s.order.GetOrder, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
		s.echoFramework.GET("/orders", s.order.GetOrders, middleware.RateLimiterWithConfig(s.getNewRateLimiterConfig()))
	} else if !s.config.Server.RateLimiterEnable {
		s.echoFramework.POST("/couriers", s.courier.AddCouriers)
		s.echoFramework.GET("/couriers/:id", s.courier.GetCourier)
		s.echoFramework.GET("/couriers", s.courier.GetCouriers)
		s.echoFramework.GET("/couriers/meta-info/:id", s.courier.GetCourierMetaInfo)

		s.echoFramework.POST("/orders", s.order.AddOrders)
		s.echoFramework.POST("/orders/complete", s.order.SetCompleteOrders)
		s.echoFramework.GET("/orders/:id", s.order.GetOrder)
		s.echoFramework.GET("/orders", s.order.GetOrders)
	}
}

func (s *Server) Start() {
	s.registerRouters()
	if err := s.echoFramework.Start(s.config.Listen.BindIP + ":" + s.config.Listen.Port); err != http.ErrServerClosed {
		s.echoFramework.Logger.Fatal(err)
	}
}

func (s *Server) GracefulShutdown() {
	defer func(clientDb *client.PostgresSql) {
		err := clientDb.Close()
		if err != nil {
			log.Info(err)
		}
	}(s.clientDb)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	sig := <-quit
	log.Info("Service is stop, got signal:", sig.String())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.echoFramework.Shutdown(ctx); err != nil {
		s.echoFramework.Logger.Fatal(err)
	}
}

func (s *Server) getNewRateLimiterConfig() middleware.RateLimiterConfig {
	var identifierExtractor func(ctx echo.Context) (string, error)
	switch s.config.Server.RateLimiterType {
	case "requests":
		identifierExtractor = func(ctx echo.Context) (string, error) {
			return "", nil
		}
	case "ip":
		identifierExtractor = func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		}
	}

	rateLimiterConfig := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(s.config.Server.RateLimiterMemoryStoreConfig.Rate),
				Burst:     s.config.Server.RateLimiterMemoryStoreConfig.Burst,
				ExpiresIn: time.Duration(s.config.Server.RateLimiterMemoryStoreConfig.ExpiresIn) * time.Second},
		),
		IdentifierExtractor: identifierExtractor,
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, struct {
			}{})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, struct {
			}{})
		},
	}
	return rateLimiterConfig
}
