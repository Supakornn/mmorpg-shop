package playerRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/models"
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
		InsertOnePlayerTransaction(pctx context.Context, req *player.PlayerTransaction) error
		GetPlayerSavingAccount(pctx context.Context, playerId string) (*player.PlayerSavingAccount, error)
		FindOnePlayerCredential(pctx context.Context, email string) (*player.Player, error)
		FindOnePlayerProfileToRefresh(pctx context.Context, playerId string) (*player.Player, error)
		GetOffset(pctx context.Context) (int64, error)
		UpsertOffset(pctx context.Context, offset int64) error
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

func (r *playerRepository) InsertOnePlayerTransaction(pctx context.Context, req *player.PlayerTransaction) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("player_transactions")

	result, err := col.InsertOne(ctx, req)
	if err != nil {
		log.Printf("error: insert one player transaction: %v", err.Error())
		return errors.New("error: insert one player transaction failed")
	}

	log.Printf("info: insert one player transaction: %v", result.InsertedID)

	return nil
}

func (r *playerRepository) GetPlayerSavingAccount(pctx context.Context, playerId string) (*player.PlayerSavingAccount, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("player_transactions")

	filter := bson.A{
		bson.D{{Key: "$match", Value: bson.D{{Key: "player_id", Value: playerId}}}},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$player_id"},
				{Key: "balance", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
			}},
		},
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "player_id", Value: "$_id"},
				{Key: "_id", Value: 0},
				{Key: "balance", Value: 1},
			}},
		},
	}

	cursor, err := col.Aggregate(ctx, filter)
	if err != nil {
		log.Printf("error: get player saving account: %v", err.Error())
		return nil, errors.New("error: get player saving account failed")
	}
	defer cursor.Close(ctx)

	result := new(player.PlayerSavingAccount)

	if cursor.Next(ctx) {
		if err := cursor.Decode(result); err != nil {
			log.Printf("error: decode player saving account: %v", err.Error())
			return nil, errors.New("error: decode player saving account failed")
		}
	} else {
		log.Printf("info: no transactions found for player: %v", playerId)
		result.PlayerId = playerId
		result.Balance = 0
	}

	return result, nil
}

func (r *playerRepository) FindOnePlayerCredential(pctx context.Context, email string) (*player.Player, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("players")

	result := new(player.Player)

	if err := col.FindOne(ctx, bson.M{"email": email}).Decode(result); err != nil {
		log.Printf("error: find one player credential: %v", err.Error())
		return nil, errors.New("error: player not found")
	}

	return result, nil
}

func (r *playerRepository) FindOnePlayerProfileToRefresh(pctx context.Context, playerId string) (*player.Player, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("players")

	result := new(player.Player)

	if err := col.FindOne(ctx, bson.M{"_id": utils.ConvertToObjectId(playerId)}).Decode(result); err != nil {
		log.Printf("error: find one player profile to refresh: %v", err.Error())
		return nil, errors.New("error: player not found")
	}

	return result, nil
}

func (r *playerRepository) GetOffset(pctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("player_transactions_queue")

	result := new(models.KafkaOffset)
	if err := col.FindOne(ctx, bson.M{}).Decode(result); err != nil {
		log.Printf("error: get offset failed: %v", err.Error())
		return -1, errors.New("error: get offset failed")
	}

	return result.Offset, nil
}

func (r *playerRepository) UpsertOffset(pctx context.Context, offset int64) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.playerDbConn(ctx)
	col := db.Collection("player_transactions_queue")

	result, err := col.UpdateOne(ctx, bson.M{}, bson.M{"$set": bson.M{"offset": offset}}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		log.Printf("error: upsert offset failed: %v", err.Error())
		return errors.New("error: upsert offset failed")
	}

	log.Printf("info: upsert offset: %v", result.ModifiedCount)

	return nil
}
