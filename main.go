package main

import (
	"context"
	"log"
	"os"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
	"github.com/Supakornn/mmorpg-shop/server"
)

func main() {
	ctx := context.Background()

	// Initialize Config
	cfg := config.LoadConfig(func() string {
		if len(os.Args) < 2 {
			log.Fatal("Error: Please provide a path to the .env file")
		}

		return os.Args[1]
	}())

	// Database Connection
	db := database.DbConn(ctx, &cfg)
	defer func() {
		if err := db.Disconnect(ctx); err != nil {
			log.Fatalf("Error: cannot disconnect from database: %s", err.Error())
		}
	}()

	// Start Server
	server.Start(ctx, &cfg, db)
}
