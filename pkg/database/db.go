package database

import (
	"context"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/config"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

func DbConn(pctx context.Context, cfg *config.Config) *mongo.Client {
	ctx, cancel := context.WithTimeout(pctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(cfg.Db.Url))
	if err != nil {
		log.Fatalf("Error: cannot connect to database: %s", err.Error())
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatalf("Error: cannot ping to database: %s", err.Error())
	}

	return client
}
