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
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	ItemRepositoryService interface {
		IsUniqueItem(pctx context.Context, title string) bool
		InsertOneItem(pctx context.Context, req *item.Item) (bson.ObjectID, error)
		FindOneItem(pctx context.Context, itemId string) (*item.Item, error)
		FindManyItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*item.ItemShowCase, error)
		CountItems(pctx context.Context, filter bson.D) (int64, error)
		UpdateOneItem(pctx context.Context, itemId string, req bson.M) error
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

func (r *itemRepository) FindManyItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*item.ItemShowCase, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	cursors, err := col.Find(ctx, filter, opts...)
	if err != nil {
		log.Printf("error: find many items: %v", err.Error())
		return make([]*item.ItemShowCase, 0), errors.New("error: find many items failed")
	}

	results := make([]*item.ItemShowCase, 0)

	for cursors.Next(ctx) {
		result := new(item.Item)
		if err := cursors.Decode(result); err != nil {
			log.Printf("error: decode item: %v", err.Error())
			return make([]*item.ItemShowCase, 0), errors.New("error: decode item failed")
		}

		results = append(results, &item.ItemShowCase{
			ItemId:   "item:" + result.Id.Hex(),
			Title:    result.Title,
			Price:    result.Price,
			ImageUrl: result.ImageUrl,
			Damage:   result.Damage,
		})
	}

	return results, nil
}

func (r *itemRepository) CountItems(pctx context.Context, filter bson.D) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	count, err := col.CountDocuments(ctx, filter)
	if err != nil {
		log.Printf("error: count items: %v", err.Error())
		return -1, errors.New("error: count items failed")
	}

	return count, nil
}

func (r *itemRepository) UpdateOneItem(pctx context.Context, itemId string, req bson.M) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.itemDbConn(ctx)
	col := db.Collection("items")

	result, err := col.UpdateOne(ctx, bson.M{"_id": utils.ConvertToObjectId(itemId)}, bson.M{"$set": req})
	if err != nil {
		log.Printf("error: update one item: %v", err.Error())
		return errors.New("error: update one item failed")
	}

	log.Printf("UpdateOneItem: %v", result.ModifiedCount)

	return nil
}
