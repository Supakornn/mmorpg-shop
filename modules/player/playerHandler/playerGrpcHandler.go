package playerHandler

import "github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"

type (
	playerGrpcHandler struct {
		playerUsecase playerUsecase.PlayerUsecaseService
	}
)

func NewPlayerGrpcHandler(playerUsecase playerUsecase.PlayerUsecaseService) playerUsecase.PlayerUsecaseService {
	return &playerGrpcHandler{playerUsecase}
}
