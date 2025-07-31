package main

import (
	"context"
	"log"
	"os"

	"github.com/Supakornn/mmorpg-shop/config"
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

	log.Println(cfg)

}
