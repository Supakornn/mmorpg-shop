package authRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/auth"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	AuthRepositoryService interface {
		InsertOneCredential(pctx context.Context, req *auth.Credential) (bson.ObjectID, error)
		CredentialSearch(pctx context.Context, grpcUrl string, req *playerPb.CredentialSearchReq) (*playerPb.PlayerProfile, error)
		FindOnePlayerCredential(pctx context.Context, credentialId string) (*auth.Credential, error)
		FindOnePlayerProfileToRefresh(pctx context.Context, grpcUrl string, req *playerPb.FindOnePlayerProfileToRefreshReq) (*playerPb.PlayerProfile, error)
		UpdateOnePlayerCredential(pctx context.Context, credentialId string, req *auth.UpdateRefreshTokenReq) error
		DeleteOnePlayerCredential(pctx context.Context, credentialId string) error
		FindOneAccessToken(pctx context.Context, accessToken string) (*auth.Credential, error)
		RoleCount(pctx context.Context) (int64, error)
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
		log.Printf("error: grpc conn failed: %v", err.Error())
		return nil, errors.New("error: grpc conn failed")
	}

	jwtauth.SetApiKeyInContext(&ctx)

	result, err := conn.Player().CredentialSearch(ctx, req)
	if err != nil {
		log.Printf("error: credential search failed: %v", err.Error())
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
		log.Printf("error: find one credential failed: %v", err.Error())
		return nil, errors.New("error: find one credential failed")
	}

	return result, nil
}

func (r *authRepository) FindOnePlayerProfileToRefresh(pctx context.Context, grpcUrl string, req *playerPb.FindOnePlayerProfileToRefreshReq) (*playerPb.PlayerProfile, error) {
	ctx, cancel := context.WithTimeout(pctx, 30*time.Second)
	defer cancel()

	conn, err := grpcconn.NewGrpcClient(grpcUrl)
	if err != nil {
		log.Printf("error: grpc conn failed: %v", err.Error())
		return nil, errors.New("error: grpc conn failed")
	}

	jwtauth.SetApiKeyInContext(&ctx)

	result, err := conn.Player().FindOnePlayerProfileToRefresh(ctx, req)
	if err != nil {
		log.Printf("error: find one player profile to refresh failed: %v", err.Error())
		return nil, errors.New("error: player not found")
	}

	return result, nil
}

func (r *authRepository) UpdateOnePlayerCredential(pctx context.Context, credentialId string, req *auth.UpdateRefreshTokenReq) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("auth")

	_, err := col.UpdateOne(
		ctx,
		bson.M{"_id": utils.ConvertToObjectId(credentialId)},
		bson.M{"$set": bson.M{
			"player_id":     req.PlayerId,
			"access_token":  req.AccessToken,
			"refresh_token": req.RefreshToken,
			"updated_at":    req.UpdatedAt,
		}},
	)
	if err != nil {
		log.Printf("error: update one player credential failed: %v", err.Error())
		return errors.New("error: update one player credential failed")
	}

	return nil
}

func (r *authRepository) DeleteOnePlayerCredential(pctx context.Context, credentialId string) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("auth")

	result, err := col.DeleteOne(ctx, bson.M{"_id": utils.ConvertToObjectId(credentialId)})
	if err != nil {
		log.Printf("Error: delete one player credential: %v", err.Error())
		return errors.New("error: delete player credential failed")
	}

	if result.DeletedCount == 0 {
		log.Printf("info: player credential not found")
		return errors.New("error: player credential not found")
	}

	log.Printf("DeleteOnePlayerCredential: %v", result.DeletedCount)

	return nil
}

func (r *authRepository) FindOneAccessToken(pctx context.Context, accessToken string) (*auth.Credential, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("auth")

	credential := new(auth.Credential)
	if err := col.FindOne(ctx, bson.M{"access_token": accessToken}).Decode(credential); err != nil {
		log.Printf("error: find one access token: %v", err.Error())
		return nil, errors.New("error: access token not found")
	}

	return credential, nil
}

func (r *authRepository) RoleCount(pctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.authDbConn(ctx)
	col := db.Collection("roles")

	count, err := col.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("error: role count failed: %v", err.Error())
		return -1, errors.New("error: roles count failed")
	}

	return count, nil
}
