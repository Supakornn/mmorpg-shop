package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryHandler"
	inventoryPb "github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryPb"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryRepository"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
)

func (s *server) inventoryService() {
	repo := inventoryRepository.NewInventoryRepository(s.db)
	usecase := inventoryUsecase.NewInventoryUsecase(repo)
	httpHandler := inventoryHandler.NewInventoryHttpHandler(s.cfg, usecase)
	grpcHandler := inventoryHandler.NewInventoryGrpcHandler(usecase)
	queueHandler := inventoryHandler.NewInventoryQueueHandler(s.cfg, usecase)

	// gRPC
	go func() {
		grpcServer, lis := grpcconn.NewGrpcServer(&s.cfg.Jwt, s.cfg.Grpc.InventoryUrl)

		inventoryPb.RegisterInventoryGrpcServiceServer(grpcServer, grpcHandler)

		s.app.Logger.Infof("Inventory gRPC server is running on %s", s.cfg.Grpc.InventoryUrl)
		grpcServer.Serve(lis)
	}()

	_ = grpcHandler
	_ = queueHandler

	inventory := s.app.Group("/inventory_v1")

	inventory.GET("", s.healthCheckService)                                                                               // Health check
	inventory.GET("/inventory/:player_id", httpHandler.FindPlayerItems, s.mid.JwtAuthorization, s.mid.PlayerIdValidation) // Find Player Items
}
