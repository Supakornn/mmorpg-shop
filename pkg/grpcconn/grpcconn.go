package grpcconn

import (
	"errors"
	"log"
	"net"

	"github.com/Supakornn/mmorpg-shop/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authPb "github.com/Supakornn/mmorpg-shop/modules/auth/authPb"
	inventoryPb "github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryPb"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
)

type (
	GrpcClientFactoryHandler interface {
		Auth() authPb.AuthGrpcServiceClient
		Player() playerPb.PlayerGrpcServiceClient
		Item() itemPb.ItemGrpcServiceClient
		Inventory() inventoryPb.InventoryGrpcServiceClient
	}

	grpcClientFactory struct {
		client *grpc.ClientConn
	}

	grpcAuth struct {
	}
)

func (g *grpcClientFactory) Auth() authPb.AuthGrpcServiceClient {
	return authPb.NewAuthGrpcServiceClient(g.client)
}

func (g *grpcClientFactory) Player() playerPb.PlayerGrpcServiceClient {
	return playerPb.NewPlayerGrpcServiceClient(g.client)
}

func (g *grpcClientFactory) Item() itemPb.ItemGrpcServiceClient {
	return itemPb.NewItemGrpcServiceClient(g.client)
}

func (g *grpcClientFactory) Inventory() inventoryPb.InventoryGrpcServiceClient {
	return inventoryPb.NewInventoryGrpcServiceClient(g.client)
}

func NewGrpcClient(host string) (GrpcClientFactoryHandler, error) {
	opts := make([]grpc.DialOption, 0)

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	clientConn, err := grpc.NewClient(host, opts...)
	if err != nil {
		log.Printf("Error: Failed to connect to GRPC: %v", err)
		return nil, errors.New("Error: Failed to connect to GRPC")
	}

	return &grpcClientFactory{
		client: clientConn,
	}, nil
}

func NewGrpcServer(cfg *config.Jwt, host string) (*grpc.Server, net.Listener) {
	opts := make([]grpc.ServerOption, 0)

	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("Error: Failed to listen GRPC on %s", host)
	}

	return grpcServer, lis
}
