package migration

import (
	"context"
	"log"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/auth"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func AuthDbConn(pctx context.Context, cfg *config.Config) *mongo.Database {
	return database.DbConn(pctx, cfg).Database("auth_db")
}

func AuthMigrate(pctx context.Context, cfg *config.Config) {
	db := AuthDbConn(pctx, cfg)
	defer db.Client().Disconnect(pctx)

	// Indexs
	// Auth
	col := db.Collection("auth")
	indexs, _ := col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "player_id", Value: 1}}},
		{Keys: bson.D{{Key: "refresh_token", Value: 1}}},
	})
	for _, index := range indexs {
		log.Printf("Index: %s created", index)
	}

	// Roles
	col = db.Collection("roles")
	indexs, _ = col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "code", Value: 1}}},
	})
	for _, index := range indexs {
		log.Printf("Index: %s created", index)
	}

	// Role Datas
	documents := func() []any {
		roles := []*auth.Role{
			{
				Title: "player",
				Code:  0,
			},
			{
				Title: "admin",
				Code:  1,
			},
		}

		docs := make([]any, 0)
		for _, r := range roles {
			docs = append(docs, r)
		}

		return docs
	}()
	results, err := col.InsertMany(pctx, documents)
	if err != nil {
		panic(err)
	}
	log.Println("Migrate auth completed", results)
}
