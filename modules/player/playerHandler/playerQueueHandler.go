package playerHandler

import (
	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"
)

type PlayerQueueHandlerService interface{}

type playerQueueHandler struct {
	cfg           *config.Config
	playerUsecase playerUsecase.PlayerUsecaseService
}

func NewPlayerQueueHandler(cfg *config.Config, playerUsecase playerUsecase.PlayerUsecaseService) PlayerQueueHandlerService {
	return &playerQueueHandler{cfg, playerUsecase}
}
