package migration

import (
	"context"
	"log"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func PlayerDbConn(pctx context.Context, cfg *config.Config) *mongo.Database {
	return database.DbConn(pctx, cfg).Database("player_db")
}

func PlayerMigrate(pctx context.Context, cfg *config.Config) {
	db := PlayerDbConn(pctx, cfg)
	defer db.Client().Disconnect(pctx)

	// Indexs
	// Player Transactions
	col := db.Collection("player_transactions")
	indexs, _ := col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "player_id", Value: 1}}},
	})

	for _, index := range indexs {
		log.Printf("index: %s created", index)
	}

	// Players
	col = db.Collection("players")
	indexs, _ = col.Indexes().CreateMany(pctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "email", Value: 1}}},
	})

	for _, index := range indexs {
		log.Printf("index: %s created", index)
	}

	// Players Datas
	documents := func() []any {
		players := []*player.Player{
			{
				Email:    "player1@supakorn.info",
				Password: "123456",
				Username: "Player1",
				PlayerRoles: []player.PlayerRole{
					{
						RoleTitle: "player",
						RoleCode:  0,
					},
				},
				CreatedAt: utils.LocalTime(),
				UpdatedAt: utils.LocalTime(),
			},
			{
				Email:    "player2@supakorn.info",
				Password: "123456",
				Username: "Player2",
				PlayerRoles: []player.PlayerRole{
					{
						RoleTitle: "player",
						RoleCode:  0,
					},
				},
				CreatedAt: utils.LocalTime(),
				UpdatedAt: utils.LocalTime(),
			},
			{
				Email:    "player3@supakorn.info",
				Password: "123456",
				Username: "Player3",
				PlayerRoles: []player.PlayerRole{
					{
						RoleTitle: "player",
						RoleCode:  0,
					},
				},
				CreatedAt: utils.LocalTime(),
				UpdatedAt: utils.LocalTime(),
			},
			{
				Email:    "admin1@supakorn.info",
				Password: "123456",
				Username: "Admin1",
				PlayerRoles: []player.PlayerRole{
					{
						RoleTitle: "player",
						RoleCode:  0,
					},
					{
						RoleTitle: "admin",
						RoleCode:  1,
					},
				},
				CreatedAt: utils.LocalTime(),
				UpdatedAt: utils.LocalTime(),
			},
		}

		docs := make([]any, 0)
		for _, p := range players {
			docs = append(docs, p)
		}

		return docs
	}()

	results, err := col.InsertMany(pctx, documents)
	if err != nil {
		panic(err)
	}
	log.Println("migrate player completed", results)

	// Player Transactions Data
	PlayerTransactions := make([]any, 0)
	for _, p := range results.InsertedIDs {
		PlayerTransactions = append(PlayerTransactions, &player.PlayerTransaction{
			PlayerId:  "player:" + p.(bson.ObjectID).Hex(),
			Amount:    1000,
			CreatedAt: utils.LocalTime(),
		})
	}

	col = db.Collection("player_transactions")
	results, err = col.InsertMany(pctx, PlayerTransactions)
	if err != nil {
		panic(err)
	}
	log.Println("migrate player transactions completed", results)

	// Player Transactions Queue
	col = db.Collection("player_transactions_queue")
	result, err := col.InsertOne(pctx, bson.M{"offset": -1})
	if err != nil {
		panic(err)
	}

	log.Println("migrate player transactions queue completed", result)
}
