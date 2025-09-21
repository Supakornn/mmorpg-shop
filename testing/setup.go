package testing

import "github.com/Supakornn/mmorpg-shop/config"

func NewTestConfig() *config.Config {
	cfg := config.LoadConfig("../env/test/.env.test")
	return &cfg
}
