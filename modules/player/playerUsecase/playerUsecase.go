package playerUsecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Supakornn/mmorpg-shop/modules/player"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerRepository"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type (
	PlayerUsecaseService interface {
		CreatePlayer(pctx context.Context, req *player.CreatePlayerReq) (*player.PlayerProfile, error)
		FindOnePlayerProfile(pctx context.Context, playerId string) (*player.PlayerProfile, error)
		AddPlayerMoney(pctx context.Context, req *player.CreatePlayerTransactionReq) (*player.PlayerSavingAccount, error)
		GetPlayerSavingAccount(pctx context.Context, playerId string) (*player.PlayerSavingAccount, error)
		FindOnePlayerCredential(pctx context.Context, email string, password string) (*playerPb.PlayerProfile, error)
	}

	playerUsecase struct {
		playerRepository playerRepository.PlayerRepositoryService
	}
)

func NewPlayerUsecase(playerRepository playerRepository.PlayerRepositoryService) PlayerUsecaseService {
	return &playerUsecase{playerRepository}
}

func (u *playerUsecase) CreatePlayer(pctx context.Context, req *player.CreatePlayerReq) (*player.PlayerProfile, error) {
	if !u.playerRepository.IsUniquePlayer(pctx, req.Email, req.Username) {
		return nil, errors.New("error: player already exists")
	}

	// Hasing Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("error: hashing password failed")
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
		return nil, errors.New("error: insert one player failed")
	}

	playerProfile, err := u.FindOnePlayerProfile(pctx, playerId.Hex())
	if err != nil {
		return nil, err
	}

	return playerProfile, nil
}

func (u *playerUsecase) FindOnePlayerProfile(pctx context.Context, playerId string) (*player.PlayerProfile, error) {
	result, err := u.playerRepository.FindOnePlayerProfile(pctx, playerId)
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")

	return &player.PlayerProfile{
		Id:        result.Id.Hex(),
		Username:  result.Username,
		Email:     result.Email,
		CreatedAt: result.CreatedAt.In(loc),
		UpdatedAt: result.UpdatedAt.In(loc),
	}, nil
}

func (u *playerUsecase) AddPlayerMoney(pctx context.Context, req *player.CreatePlayerTransactionReq) (*player.PlayerSavingAccount, error) {
	if err := u.playerRepository.InsertOnePlayerTransaction(pctx, &player.PlayerTransaction{
		PlayerId:  req.PlayerId,
		Amount:    req.Amount,
		CreatedAt: utils.LocalTime(),
	}); err != nil {
		return nil, err
	}

	return u.GetPlayerSavingAccount(pctx, req.PlayerId)
}

func (u *playerUsecase) GetPlayerSavingAccount(pctx context.Context, playerId string) (*player.PlayerSavingAccount, error) {
	result, err := u.playerRepository.GetPlayerSavingAccount(pctx, playerId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (u *playerUsecase) FindOnePlayerCredential(pctx context.Context, email string, password string) (*playerPb.PlayerProfile, error) {
	result, err := u.playerRepository.FindOnePlayerCredential(pctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(password)); err != nil {
		log.Printf("error: password not match: %v", err)
		return nil, errors.New("error: password is incorrect")
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")

	roleCode := 0
	for _, v := range result.PlayerRoles {
		roleCode = v.RoleCode
	}

	return &playerPb.PlayerProfile{
		Id:        result.Id.Hex(),
		Email:     result.Email,
		Username:  result.Username,
		RoleCode:  int32(roleCode),
		CreatedAt: result.CreatedAt.In(loc).String(),
		UpdatedAt: result.UpdatedAt.In(loc).String(),
	}, nil
}
