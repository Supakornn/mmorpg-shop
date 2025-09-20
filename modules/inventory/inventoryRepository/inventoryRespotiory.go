package inventoryRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	InventoryRepositoryService interface {
		FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error)
		FindPlayerItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*inventory.Inventory, error)
		CountPlayerItems(pctx context.Context, playerId string) (int64, error)
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
