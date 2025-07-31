package main

import (
	"context"
	"log"
	"os"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/pkg/database"
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
	defer db.Disconnect(ctx)
}
