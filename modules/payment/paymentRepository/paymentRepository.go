package paymentRepository

import (
	"context"
	"errors"
	"log"
	"time"

	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/models"
	"github.com/Supakornn/mmorpg-shop/pkg/grpcconn"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	PaymentRepositoryService interface {
		FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error)
		GetOffset(pctx context.Context) (int64, error)
		UpsertOffset(pctx context.Context, offset int64) error
	}

	paymentRepository struct {
		db *mongo.Client
	}
)

func NewPaymentRepository(db *mongo.Client) PaymentRepositoryService {
	return &paymentRepository{db}
}

func (r *paymentRepository) paymentDbConn(pctx context.Context) *mongo.Database {
	return r.db.Database("payment_db")
}

func (r *paymentRepository) FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error) {
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

func (r *paymentRepository) GetOffset(pctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.paymentDbConn(ctx)
	col := db.Collection("payment_queue")

	result := new(models.KafkaOffset)
	if err := col.FindOne(ctx, bson.M{}).Decode(result); err != nil {
		log.Printf("error: get offset failed: %v", err.Error())
		return -1, errors.New("error: get offset failed")
	}

	return result.Offset, nil
}

func (r *paymentRepository) UpsertOffset(pctx context.Context, offset int64) error {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	db := r.paymentDbConn(ctx)
	col := db.Collection("payment_queue")

	result, err := col.UpdateOne(ctx, bson.M{}, bson.M{"$set": bson.M{"offset": offset}}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		log.Printf("error: upsert offset failed: %v", err.Error())
		return errors.New("error: upsert offset failed")
	}

	log.Printf("info: upsert offset: %v", result.ModifiedCount)

	return nil
}
