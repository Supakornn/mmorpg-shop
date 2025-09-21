package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryHandler"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryRepository"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryUsecase"
)

func (s *server) inventoryService() {
	repo := inventoryRepository.NewInventoryRepository(s.db)
	usecase := inventoryUsecase.NewInventoryUsecase(repo)
	httpHandler := inventoryHandler.NewInventoryHttpHandler(s.cfg, usecase)
	queueHandler := inventoryHandler.NewInventoryQueueHandler(s.cfg, usecase)

	_ = queueHandler

	inventory := s.app.Group("/inventory_v1")

	inventory.GET("", s.healthCheckService)                                                                               // Health check
	inventory.GET("/inventory/:player_id", httpHandler.FindPlayerItems, s.mid.JwtAuthorization, s.mid.PlayerIdValidation) // Find Player Items
}
