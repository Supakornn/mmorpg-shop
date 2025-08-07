package playerRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	PlayerRepositoryService interface {
		IsUniquePlayer(pctx context.Context, email, username string) bool
		InsertOnePlayer(pctx context.Context, req *player.Player) (bson.ObjectID, error)
		FindOnePlayerProfile(pctx context.Context, playerId string) (*player.PlayerProfileBson, error)
	}

	playerRepository struct {
		db *mongo.Client
	}
)

func NewPlayerRepository(db *mongo.Client) PlayerRepositoryService {
	return &playerRepository{db}
}

func (r *playerRepository) playerDbConn(pctx context.Context) *mongo.Database {
	return r.db.Database("player_db")
}

func (r *playerRepository) IsUniquePlayer(pctx context.Context, email, username string) bool {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("players")

	player := new(player.Player)
	if err := col.FindOne(
		ctx,
		bson.M{"$or": []bson.M{
			{"username": username},
			{"email": email},
		}},
	).Decode(player); err != nil {
		log.Printf("error: is unique player: %v", err.Error())
		return true
	}
	return false
}

func (r *playerRepository) InsertOnePlayer(pctx context.Context, req *player.Player) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("players")

	playerId, err := col.InsertOne(ctx, req)
	if err != nil {
		log.Printf("error: insert one player: %v", err.Error())
		return bson.NilObjectID, errors.New("error: insert one player failed")
	}

	return playerId.InsertedID.(bson.ObjectID), nil
}

func (r *playerRepository) FindOnePlayerProfile(pctx context.Context, playerId string) (*player.PlayerProfileBson, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("players")

	// Validate ObjectID format
	objectId := utils.ConvertToObjectId(playerId)
	if objectId == bson.NilObjectID {
		log.Printf("error: invalid player id format: %v", playerId)
		return nil, errors.New("error: invalid player id format")
	}

	result := new(player.PlayerProfileBson)

	if err := col.FindOne(
		ctx,
		bson.M{"_id": objectId},
		options.FindOne().SetProjection(bson.M{
			"_id":        1,
			"username":   1,
			"email":      1,
			"created_at": 1,
			"updated_at": 1,
		}),
	).Decode(result); err != nil {
		log.Printf("error: find one player profile: %v", err.Error())
		return nil, errors.New("error: player not found")
	}

	return result, nil
}
