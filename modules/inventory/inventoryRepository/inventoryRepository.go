package inventoryRepository

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/models"
	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/Supakornn/mmorpg-shop/pkg/queue"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	InventoryRepositoryService interface {
		FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error)
		FindPlayerItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*inventory.Inventory, error)
		CountPlayerItems(pctx context.Context, playerId string) (int64, error)
		GetOffset(pctx context.Context) (int64, error)
		UpsertOffset(pctx context.Context, offset int64) error
		AddPlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error
		RemovePlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error
		InsertOnePlayerItem(pctx context.Context, req *inventory.Inventory) (bson.ObjectID, error)
		FindOnePlayerItem(pctx context.Context, playerId, itemId string) bool
		DeleteOneInventory(pctx context.Context, inventoryId string) error
		DeleteOnePlayerItem(pctx context.Context, playerId, itemId string) error
	}

	inventoryRepository struct {
		db *mongo.Client
	}
)

func NewInventoryRepository(db *mongo.Client) InventoryRepositoryService {
	return &inventoryRepository{db}
}

func (r *inventoryRepository) inventoryDbConn(pctx context.Context) *mongo.Database {
	return r.db.Database("inventory_db")
}

func (r *inventoryRepository) FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error) {
	ctx, cancel := context.WithTimeout(pctx, 30*time.Second)
	defer cancel()

	conn, err := grpcconn.NewGrpcClient(grpcUrl)
	if err != nil {
		log.Printf("error: grpc conn failed: %v", err.Error())
		return nil, errors.New("error: grpc conn failed")
	}

	jwtauth.SetApiKeyInContext(&ctx)

	result, err := conn.Item().FindItemsInIds(ctx, req)
	if err != nil {
		log.Printf("error: find items in ids failed: %v", err.Error())
		return nil, errors.New("error: find items in ids failed")
	}

	if result == nil {
		return nil, errors.New("error: items not found")
	}

	if result.Items == nil {
		return nil, errors.New("error: items not found")
	}

	if len(result.Items) == 0 {
		return nil, errors.New("error: items not found")
	}

	return result, nil
}

func (r *inventoryRepository) FindPlayerItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*inventory.Inventory, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	cursors, err := col.Find(ctx, filter, opts...)
	if err != nil {
		log.Printf("error: find player items failed: %v", err.Error())
		return nil, errors.New("error: find player items failed")
	}

	results := make([]*inventory.Inventory, 0)

	for cursors.Next(ctx) {
		result := new(inventory.Inventory)
		if err := cursors.Decode(result); err != nil {
			log.Printf("error: decode player items failed: %v", err.Error())
			return nil, errors.New("error: decode player items failed")
		}

		results = append(results, result)
	}

	return results, nil
}

func (r *inventoryRepository) CountPlayerItems(pctx context.Context, playerId string) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	log.Println("playerId", playerId)

	count, err := col.CountDocuments(ctx, bson.M{"player_id": playerId})
	if err != nil {
		log.Printf("error: count player items: %v", err.Error())
		return -1, errors.New("error: count player items failed")
	}

	log.Println("count", count)

	return count, nil
}

func (r *inventoryRepository) FindOnePlayerItem(pctx context.Context, playerId, itemId string) bool {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	result := new(inventory.Inventory)
	if err := col.FindOne(ctx, bson.M{"player_id": playerId, "item_id": itemId}).Decode(result); err != nil {
		log.Printf("error: find one player item: %v", err.Error())
		return false
	}

	return true
}

func (r *inventoryRepository) GetOffset(pctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("player_inventory_queue")

	result := new(models.KafkaOffset)
	if err := col.FindOne(ctx, bson.M{}).Decode(result); err != nil {
		log.Printf("error: get offset failed: %v", err.Error())
		return -1, errors.New("error: get offset failed")
	}

	return result.Offset, nil
}

func (r *inventoryRepository) UpsertOffset(pctx context.Context, offset int64) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("player_inventory_queue")

	result, err := col.UpdateOne(ctx, bson.M{}, bson.M{"$set": bson.M{"offset": offset}}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		log.Printf("error: upsert offset failed: %v", err.Error())
		return errors.New("error: upsert offset failed")
	}

	log.Printf("info: upsert offset: %v", result.ModifiedCount)

	return nil
}

func (r *inventoryRepository) AddPlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	reqInBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error: marshal request failed: %v", err.Error())
		return errors.New("error: marshal request failed")
	}

	if err := queue.PushMessageWithKeyToQueue([]string{cfg.Kafka.Url}, cfg.Kafka.ApiKey, cfg.Kafka.Secret, "payment", "buy", reqInBytes); err != nil {
		log.Printf("Error: push message with key to queue failed: %v", err.Error())
		return errors.New("error: push message with key to queue failed")
	}

	return nil
}

func (r *inventoryRepository) RemovePlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	reqInBytes, err := json.Marshal(req)
	if err != nil {
		log.Printf("Error: marshal request failed: %v", err.Error())
		return errors.New("error: marshal request failed")
	}

	if err := queue.PushMessageWithKeyToQueue([]string{cfg.Kafka.Url}, cfg.Kafka.ApiKey, cfg.Kafka.Secret, "payment", "sell", reqInBytes); err != nil {
		log.Printf("Error: push message with key to queue failed: %v", err.Error())
		return errors.New("error: push message with key to queue failed")
	}

	return nil
}

func (r *inventoryRepository) DeleteOnePlayerItem(pctx context.Context, playerId, itemId string) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	result, err := col.DeleteOne(ctx, bson.M{"player_id": playerId, "item_id": itemId})
	if err != nil {
		log.Printf("error: delete one player item: %v", err.Error())
		return errors.New("error: delete one player item failed")
	}

	log.Printf("delete one player item: %v", result.DeletedCount)

	return nil
}

func (r *inventoryRepository) InsertOnePlayerItem(pctx context.Context, req *inventory.Inventory) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	result, err := col.InsertOne(ctx, req)
	if err != nil {
		log.Printf("error: insert one player item: %v", err.Error())
		return bson.NilObjectID, errors.New("error: insert one player item failed")
	}

	return result.InsertedID.(bson.ObjectID), nil
}

func (r *inventoryRepository) DeleteOneInventory(pctx context.Context, inventoryId string) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.inventoryDbConn(ctx)
	col := db.Collection("inventories")

	result, err := col.DeleteOne(ctx, bson.M{"_id": utils.ConvertToObjectId(inventoryId)})
	if err != nil {
		log.Printf("error: delete one inventory: %v", err.Error())
		return errors.New("error: delete one inventory failed")
	}

	log.Printf("delete one inventory: %v", result.DeletedCount)

	return nil
}
