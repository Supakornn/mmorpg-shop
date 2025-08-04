package main

import (
	"context"
	"log"
	"os"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/pkg/database/migration"
)

func main() {
	ctx := context.Background()
	_ = ctx

	// Initialize Config
	cfg := config.LoadConfig(func() string {
		if len(os.Args) < 2 {
			log.Fatal("Error: Please provide a path to the .env file")
		}

		return os.Args[1]
	}())

	switch cfg.App.Name {
	case "auth":
		migration.AuthMigrate(ctx, &cfg)
	case "player":
	case "item":
	case "inventory":
	}
}
