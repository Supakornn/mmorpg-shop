package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/player/playerHandler"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerRepository"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
)

func (s *server) playerService() {
	repo := playerRepository.NewPlayerRepository(s.db)
	usecase := playerUsecase.NewPlayerUsecase(repo)
	httpHandler := playerHandler.NewPlayerHttpHandler(s.cfg, usecase)
	grpcHandler := playerHandler.NewPlayerGrpcHandler(usecase)
	queueHandler := playerHandler.NewPlayerQueueHandler(s.cfg, usecase)

	go queueHandler.DockedPlayerMoney()
	go queueHandler.RollbackPlayerTransaction()

	// gRPC
	go func() {
		grpcServer, lis := grpcconn.NewGrpcServer(&s.cfg.Jwt, s.cfg.Grpc.PlayerUrl)

		playerPb.RegisterPlayerGrpcServiceServer(grpcServer, grpcHandler)

		s.app.Logger.Infof("Player gRPC server is running on %s", s.cfg.Grpc.PlayerUrl)
		grpcServer.Serve(lis)
	}()

	// Routes
	player := s.app.Group("/player_v1")

	player.GET("", s.healthCheckService)                                                                        // Health check
	player.POST("/player/register", httpHandler.CreatePlayer)                                                   // Create Player
	player.GET("/player/:player_id", httpHandler.FindOnePlayerProfile)                                          // Find One Player Profile
	player.POST("/player/add-money", httpHandler.AddPlayerMoney, s.mid.JwtAuthorization)                        // Add Player Money
	player.GET("/player/saving-account/my-account", httpHandler.GetPlayerSavingAccount, s.mid.JwtAuthorization) // Get Player Saving Account
}
