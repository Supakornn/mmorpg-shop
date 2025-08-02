package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/player/playerHandler"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerRepository"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"
)

func (s *server) playerService() {
	repo := playerRepository.NewPlayerRepository(s.db)
	usecase := playerUsecase.NewPlayerUsecase(repo)
	httpHandler := playerHandler.NewPlayerHttpHandler(s.cfg, usecase)
	grpcHandler := playerHandler.NewPlayerGrpcHandler(usecase)
	queueHandler := playerHandler.NewPlayerQueueHandler(s.cfg, usecase)

	_ = httpHandler
	_ = grpcHandler
	_ = queueHandler

	player := s.app.Group("/player_v1")

	// Health check
	_ = player
}
