package authRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/auth"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	AuthRepositoryService interface {
		InsertOneCredential(pctx context.Context, req *auth.Credential) (bson.ObjectID, error)
		CredentialSearch(pctx context.Context, grpcUrl string, req *playerPb.CredentialSearchReq) (*playerPb.PlayerProfile, error)
	}

	authRepository struct {
		db *mongo.Client
	}
)

func NewAuthRepository(db *mongo.Client) AuthRepositoryService {
	return &authRepository{db}
}

func (r *authRepository) authDbConn(pctx context.Context) *mongo.Database {
	return r.db.Database("auth_db")
}

func (r *authRepository) CredentialSearch(pctx context.Context, grpcUrl string, req *playerPb.CredentialSearchReq) (*playerPb.PlayerProfile, error) {
	ctx, cancel := context.WithTimeout(pctx, 30*time.Second)
	defer cancel()

	conn, err := grpcconn.NewGrpcClient(grpcUrl)
	if err != nil {
		log.Printf("error: grpc conn failed: %v", err)
		return nil, errors.New("error: grpc conn failed")
	}

	result, err := conn.Player().CredentialSearch(ctx, req)
	if err != nil {
		log.Printf("error: credential search failed: %v", err)
		return nil, errors.New("error: credential search failed")
	}

	return result, nil
}

func (r *authRepository) InsertOneCredential(pctx context.Context, req *auth.Credential) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("auth")

	result, err := col.InsertOne(ctx, req)
	if err != nil {
		log.Printf("error: insert credential failed: %v", err.Error())
		return bson.NewObjectID(), errors.New("error: insert credential failed")
	}

	return result.InsertedID.(bson.ObjectID), nil
}
