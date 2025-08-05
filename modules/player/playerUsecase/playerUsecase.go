package playerUsecase

import (
	"context"
	"errors"

	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerRepository"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type (
	PlayerUsecaseService interface {
		CreatePlayer(pctx context.Context, req *player.CreatePlayerReq) (string, error)
	}

	playerUsecase struct {
		playerRepository playerRepository.PlayerRepositoryService
	}
)

func NewPlayerUsecase(playerRepository playerRepository.PlayerRepositoryService) PlayerUsecaseService {
	return &playerUsecase{playerRepository}
}

func (u *playerUsecase) CreatePlayer(pctx context.Context, req *player.CreatePlayerReq) (string, error) {
	if !u.playerRepository.IsUniquePlayer(pctx, req.Email, req.Username) {
		return "", errors.New("error: player already exists")
	}

	// Hasing Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("error: hashing password failed")
	}

	// Insert One Player
	playerId, err := u.playerRepository.InsertOnePlayer(pctx, &player.Player{
		Email:     req.Email,
		Password:  string(hashedPassword),
		Username:  req.Username,
		CreatedAt: utils.LocalTime(),
		UpdatedAt: utils.LocalTime(),
		PlayerRoles: []player.PlayerRole{
			{
				RoleTitle: "Player",
				RoleCode:  0,
			},
		},
	})

	if err != nil {
		return "", errors.New("error: insert one player failed")
	}

	return playerId.Hex(), nil
}
