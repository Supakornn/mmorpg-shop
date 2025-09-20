package itemRepository

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type (
	ItemRepositoryService interface {
		IsUniqueItem(pctx context.Context, title string) bool
		InsertOneItem(pctx context.Context, req *item.Item) (bson.ObjectID, error)
		FindOneItem(pctx context.Context, itemId string) (*item.Item, error)
	}

	itemRepository struct {
		db *mongo.Client
	}
)

func NewItemRepository(db *mongo.Client) ItemRepositoryService {
	return &itemRepository{db}
}

func (r *itemRepository) itemDbConn(pctx context.Context) *mongo.Database {
	return r.db.Database("item_db")
}

func (r *itemRepository) IsUniqueItem(pctx context.Context, title string) bool {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	item := new(item.Item)
	if err := col.FindOne(ctx, bson.M{"title": title}).Decode(item); err != nil {
		log.Printf("error: is unique item: %v", err.Error())
		return true
	}
	return false
}

func (r *itemRepository) InsertOneItem(pctx context.Context, req *item.Item) (bson.ObjectID, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	itemId, err := col.InsertOne(ctx, req)
	if err != nil {
		log.Printf("error: insert one item: %v", err.Error())
		return bson.NilObjectID, errors.New("error: insert one item failed")
	}

	return itemId.InsertedID.(bson.ObjectID), nil
}

func (r *itemRepository) FindOneItem(pctx context.Context, itemId string) (*item.Item, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	result := new(item.Item)
	if err := col.FindOne(ctx, bson.M{"_id": utils.ConvertToObjectId(itemId)}).Decode(result); err != nil {
		log.Printf("error: find one item: %v", err.Error())
		return nil, errors.New("error: item not found")
	}

	return result, nil
}
