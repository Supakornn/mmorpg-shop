package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/middleware/middlewareHandler"
	"github.com/Supakornn/mmorpg-shop/modules/middleware/middlewareRepository"
	"github.com/Supakornn/mmorpg-shop/modules/middleware/middlewareUsecase"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	server struct {
		app *echo.Echo
		db  *mongo.Client
		cfg *config.Config
		mid middlewareHandler.MiddlewareHandlerService
	}
)

func newMiddleware(cfg *config.Config) middlewareHandler.MiddlewareHandlerService {
	repo := middlewareRepository.NewMiddlewareRepository()
	usecase := middlewareUsecase.NewMiddlewareUsecase(repo)
	return middlewareHandler.NewMiddlewareHandler(cfg, usecase)
}

func (s *server) graceFullShutdown(pctx context.Context, quit <-chan os.Signal) {
	log.Printf("start service: %s", s.cfg.App.Name)

	<-quit

	log.Printf("shutting down service: %s", s.cfg.App.Name)

	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	if err := s.app.Shutdown(ctx); err != nil {
		s.app.Logger.Fatalf("error: %s", err.Error())
	}

	log.Printf("service: %s shutdown complete", s.cfg.App.Name)
}

func (s *server) httpListening() {
	if err := s.app.Start(s.cfg.App.Url); err != nil && err != http.ErrServerClosed {
		s.app.Logger.Fatalf("error: %s", err.Error())
	}
}

func Start(pctx context.Context, cfg *config.Config, db *mongo.Client) {
	s := &server{
		app: echo.New(),
		db:  db,
		cfg: cfg,
		mid: newMiddleware(cfg),
	}

	// Middleware Basic
	// Request Timeout
	s.app.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "error: request timeout",
		Timeout:      30 * time.Second,
	}))

	// CORS
	s.app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE},
	}))

	// Body Limit
	s.app.Use(middleware.BodyLimit("10M"))
	switch s.cfg.App.Name {
	case "auth":
		s.authService()
	case "player":
		s.playerService()
	case "item":
		s.itemService()
	case "inventory":
		s.inventoryService()
	case "payment":
		s.paymentService()
	}

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	s.app.Use(middleware.Logger())

	go s.graceFullShutdown(pctx, quit)

	// Listening
	s.httpListening()
}
