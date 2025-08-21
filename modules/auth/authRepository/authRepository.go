package authRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/auth"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	AuthRepositoryService interface {
		InsertOneCredential(pctx context.Context, req *auth.Credential) (bson.ObjectID, error)
		CredentialSearch(pctx context.Context, grpcUrl string, req *playerPb.CredentialSearchReq) (*playerPb.PlayerProfile, error)
		FindOnePlayerCredential(pctx context.Context, credentialId string) (*auth.Credential, error)
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
		return nil, errors.New("error: email or password is incorrect")
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

func (r *authRepository) FindOnePlayerCredential(pctx context.Context, credentialId string) (*auth.Credential, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("auth")

	result := new(auth.Credential)

	if err := col.FindOne(ctx, bson.M{"_id": utils.ConvertToObjectId(credentialId)}).Decode(result); err != nil {
		log.Printf("error: find one credential failed: %v", err)
		return nil, errors.New("error: find one credential failed")
	}

	return result, nil
}
