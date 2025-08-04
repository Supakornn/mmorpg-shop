package server

import (
	"github.com/Supakornn/mmorpg-shop/modules/item/itemHandler"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemRepository"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
)

func (s *server) itemService() {
	repo := itemRepository.NewItemRepository(s.db)
	usecase := itemUsecase.NewItemUsecase(repo)
	httpHandler := itemHandler.NewItemHttpHandler(s.cfg, usecase)
	grpcHandler := itemHandler.NewItemGrpcHandler(usecase)

	// gRPC
	go func() {
		grpcServer, lis := grpcconn.NewGrpcServer(&s.cfg.Jwt, s.cfg.Grpc.ItemUrl)

		itemPb.RegisterItemGrpcServiceServer(grpcServer, grpcHandler)

		s.app.Logger.Infof("Item gRPC server is running on %s", s.cfg.Grpc.ItemUrl)
		grpcServer.Serve(lis)
	}()

	_ = httpHandler
	_ = grpcHandler

	item := s.app.Group("/item_v1")

	// Health check
	item.GET("", s.healthCheckService)
}
