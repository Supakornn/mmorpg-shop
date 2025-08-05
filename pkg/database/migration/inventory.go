package migration

import (
	"context"
	"log"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func InventoryDbConn(pctx context.Context, cfg *config.Config) *mongo.Database {
	return database.DbConn(pctx, cfg).Database("inventory_db")
}

func InventoryMigrate(pctx context.Context, cfg *config.Config) {
	db := InventoryDbConn(pctx, cfg)
	defer db.Client().Disconnect(pctx)

	// Indexs
	// Inventory
	col := db.Collection("inventories")
	indexs, _ := col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "player_id", Value: 1}, {Key: "item_id", Value: 1}}},
	})

	for _, index := range indexs {
		log.Printf("Index: %s created", index)
	}

	col = db.Collection("player_inventory_queue")

	results, err := col.InsertOne(pctx, bson.M{"offset": -1})
	if err != nil {
		panic(err)
	}

	log.Println("Migrate inventory completed", results)
}
