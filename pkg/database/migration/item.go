package migration

import (
	"context"
	"log"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func ItemDbConn(pctx context.Context, cfg *config.Config) *mongo.Database {
	return database.DbConn(pctx, cfg).Database("item_db")
}

func ItemMigrate(pctx context.Context, cfg *config.Config) {
	db := ItemDbConn(pctx, cfg)
	defer db.Client().Disconnect(pctx)

	// Indexs
	// Item
	col := db.Collection("items")
	indexs, _ := col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "title", Value: 1}}},
	})
	for _, index := range indexs {
		log.Printf("Index: %s created", index)
	}

	// Items Datas
	documents := func() []any {
		items := []*item.Item{
			{
				Title:       "Sword",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/sword.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Shield",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/shield.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Helmet",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/helmet.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Armor",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/armor.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Boots",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/boots.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Gloves",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/gloves.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
			{
				Title:       "Ring",
				Price:       100,
				Damage:      10,
				ImageUrl:    "https://example.com/ring.png",
				UsageStatus: true,
				CreatedAt:   utils.LocalTime(),
				UpdatedAt:   utils.LocalTime(),
			},
		}
		docs := make([]any, 0)
		for _, i := range items {
			docs = append(docs, i)
		}

		return docs
	}()
	results, err := col.InsertMany(pctx, documents)
	if err != nil {
		panic(err)
	}
	log.Println("Migrate item completed", results)
}
