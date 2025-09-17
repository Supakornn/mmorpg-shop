package middlewareRepository

import (
	"context"
	"errors"
	"log"
	"time"

	authPb "github.com/Supakornn/mmorpg-shop/modules/auth/authPb"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
)

type (
	MiddlewareRepositoryService interface {
		AccessTokenSearch(pctx context.Context, grpcUrl, accessToken string) error
	}

	middlewareRepository struct{}
)

func NewMiddlewareRepository() MiddlewareRepositoryService {
	return &middlewareRepository{}
}

func (r *middlewareRepository) AccessTokenSearch(pctx context.Context, grpcUrl, accessToken string) error {
	ctx, cancel := context.WithTimeout(pctx, 30*time.Second)
	defer cancel()

	conn, err := grpcconn.NewGrpcClient(grpcUrl)
	if err != nil {
		log.Printf("error: grpc conn failed: %v", err.Error())
		return errors.New("error: grpc conn failed")
	}

	result, err := conn.Auth().AccessTokenSearch(ctx, &authPb.AccessTokenSearchReq{
		AccessToken: accessToken,
	})
	if err != nil {
		log.Printf("error: access token search failed: %v", err.Error())
		return errors.New("error: access token is invalid")
	}

	if !result.IsValid {
		log.Printf("error: access token is invalid")
		return errors.New("error: access token is invalid")
	}

	return nil
}
