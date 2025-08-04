package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/auth/authHandler"
	authPb "github.com/Supakornn/mmorpg-shop/modules/auth/authPb"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authRepository"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
)

func (s *server) authService() {
	repo := authRepository.NewAuthRepository(s.db)
	usecase := authUsecase.NewAuthUsecase(repo)
	httpHandler := authHandler.NewAuthHttpHandler(s.cfg, usecase)
	grpcHandler := authHandler.NewAuthGrpcHandler(usecase)

	// gRPC
	go func() {
		grpcServer, lis := grpcconn.NewGrpcServer(&s.cfg.Jwt, s.cfg.Grpc.AuthUrl)

		authPb.RegisterAuthGrpcServiceServer(grpcServer, grpcHandler)

		s.app.Logger.Infof("Auth gRPC server is running on %s", s.cfg.Grpc.AuthUrl)
		grpcServer.Serve(lis)
	}()

	_ = httpHandler
	_ = grpcHandler

	auth := s.app.Group("/auth_v1")

	// Health check
	auth.GET("", s.healthCheckService)
}
