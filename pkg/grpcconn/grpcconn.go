package grpcconn

import (
	"context"
	"errors"
	"log"
	"net"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

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
		secretKey string
	}
)

func (g *grpcAuth) unaryAuthorization(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Printf("error: metadata is not provided")
		return nil, errors.New("error: metadata is not provided")
	}

	authHeader, ok := md["auth"]
	if !ok {
		log.Printf("error: auth header is not provided")
		return nil, errors.New("error: auth header is not provided")
	}

	if len(authHeader) == 0 {
		log.Printf("error: auth header is not provided")
		return nil, errors.New("error: auth header is not provided")
	}

	claims, err := jwtauth.ParseToken(g.secretKey, authHeader[0])
	if err != nil {
		log.Printf("error: failed to parse token: %v", err)
		return nil, errors.New("error: failed to parse token")
	}

	log.Printf("claims: %v", claims)

	return handler(ctx, req)
}

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
		log.Printf("error: failed to connect to grpc: %v", err)
		return nil, errors.New("error: failed to connect to grpc")
	}

	return &grpcClientFactory{
		client: clientConn,
	}, nil
}

func NewGrpcServer(cfg *config.Jwt, host string) (*grpc.Server, net.Listener) {
	opts := make([]grpc.ServerOption, 0)

	grpcAuth := &grpcAuth{
		secretKey: cfg.ApiSecretKey,
	}

	opts = append(opts, grpc.UnaryInterceptor(grpcAuth.unaryAuthorization))

	grpcServer := grpc.NewServer(opts...)

	lis, err := net.Listen("tcp", host)
	if err != nil {
		log.Fatalf("error: failed to listen grpc on %s", host)
	}

	return grpcServer, lis
}
