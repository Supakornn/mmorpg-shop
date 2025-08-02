package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/auth/authHandler"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authRepository"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authUsecase"
)

func (s *server) authService() {
	repo := authRepository.NewAuthRepository(s.db)
	usecase := authUsecase.NewAuthUsecase(repo)
	httpHandler := authHandler.NewAuthHttpHandler(s.cfg, usecase)
	grpcHandler := authHandler.NewAuthGrpcHandler(usecase)

	_ = httpHandler
	_ = grpcHandler

	auth := s.app.Group("/auth_v1")

	// Health check
	auth.GET("", s.healthCheckService)
}
