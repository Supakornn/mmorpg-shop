package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/Supakornn/mmorpg-shop/modules/player"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerRepository"
	"github.com/Supakornn/mmorpg-shop/modules/player/playerUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

type (
	testCreatePlayer struct {
		name     string
		ctx      context.Context
		req      *player.CreatePlayerReq
		expected *player.PlayerProfile
		isErr    bool
	}

	testFindOnePlayerProfile struct {
		name     string
		ctx      context.Context
		playerId string
		expected *player.PlayerProfile
		isErr    bool
	}

	testAddPlayerMoney struct {
		name     string
		ctx      context.Context
		req      *player.CreatePlayerTransactionReq
		expected *player.PlayerSavingAccount
		isErr    bool
	}

	testGetPlayerSavingAccount struct {
		name     string
		ctx      context.Context
		playerId string
		expected *player.PlayerSavingAccount
		isErr    bool
	}

	testFindOnePlayerCredential struct {
		name     string
		ctx      context.Context
		email    string
		password string
		expected *playerPb.PlayerProfile
		isErr    bool
	}
)

func TestCreatePlayer(t *testing.T) {
	repoMock := new(playerRepository.PlayerRepositoryMock)
	usecase := playerUsecase.NewPlayerUsecase(repoMock)

	ctx := context.Background()
	playerId := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testCreatePlayer{
		{
			name: "success create player",
			ctx:  ctx,
			req: &player.CreatePlayerReq{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser",
			},
			expected: &player.PlayerProfile{
				Id:        playerId.Hex(),
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			isErr: false,
		},
		{
			name: "failed create player - user already exists",
			ctx:  ctx,
			req: &player.CreatePlayerReq{
				Email:    "existing@example.com",
				Password: "password123",
				Username: "existinguser",
			},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("IsUniquePlayer", ctx, "test@example.com", "testuser").Return(true)
	repoMock.On("InsertOnePlayer", ctx, mock.AnythingOfType("*player.Player")).Return(playerId, nil)
	repoMock.On("FindOnePlayerProfile", ctx, playerId.Hex()).Return(&player.PlayerProfileBson{
		Id:        playerId,
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}, nil)

	// Failed case
	repoMock.On("IsUniquePlayer", ctx, "existing@example.com", "existinguser").Return(false)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.CreatePlayer(test.ctx, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.Id, result.Id)
				assert.Equal(t, test.expected.Email, result.Email)
				assert.Equal(t, test.expected.Username, result.Username)
			}
		})
	}
}

func TestFindOnePlayerProfile(t *testing.T) {
	repoMock := new(playerRepository.PlayerRepositoryMock)
	usecase := playerUsecase.NewPlayerUsecase(repoMock)

	ctx := context.Background()
	playerId := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testFindOnePlayerProfile{
		{
			name:     "success find player profile",
			ctx:      ctx,
			playerId: playerId.Hex(),
			expected: &player.PlayerProfile{
				Id:        playerId.Hex(),
				Email:     "test@example.com",
				Username:  "testuser",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
			isErr: false,
		},
		{
			name:     "failed find player profile - not found",
			ctx:      ctx,
			playerId: "invalid_player_id",
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("FindOnePlayerProfile", ctx, playerId.Hex()).Return(&player.PlayerProfileBson{
		Id:        playerId,
		Email:     "test@example.com",
		Username:  "testuser",
		CreatedAt: testTime,
		UpdatedAt: testTime,
	}, nil)

	// Failed case
	repoMock.On("FindOnePlayerProfile", ctx, "invalid_player_id").Return(&player.PlayerProfileBson{}, errors.New("player not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.FindOnePlayerProfile(test.ctx, test.playerId)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.Id, result.Id)
				assert.Equal(t, test.expected.Email, result.Email)
				assert.Equal(t, test.expected.Username, result.Username)
			}
		})
	}
}

func TestAddPlayerMoney(t *testing.T) {
	repoMock := new(playerRepository.PlayerRepositoryMock)
	usecase := playerUsecase.NewPlayerUsecase(repoMock)

	ctx := context.Background()
	transactionId := bson.NewObjectID()

	tests := []testAddPlayerMoney{
		{
			name: "success add player money",
			ctx:  ctx,
			req: &player.CreatePlayerTransactionReq{
				PlayerId: "player:001",
				Amount:   100.0,
			},
			expected: &player.PlayerSavingAccount{
				PlayerId: "player:001",
				Balance:  100.0,
			},
			isErr: false,
		},
		{
			name: "failed add player money - transaction failed",
			ctx:  ctx,
			req: &player.CreatePlayerTransactionReq{
				PlayerId: "invalid_player",
				Amount:   100.0,
			},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("InsertOnePlayerTransaction", ctx, mock.AnythingOfType("*player.PlayerTransaction")).Return(transactionId, nil).Once()
	repoMock.On("GetPlayerSavingAccount", ctx, "player:001").Return(&player.PlayerSavingAccount{
		PlayerId: "player:001",
		Balance:  100.0,
	}, nil).Once()

	// Failed case
	repoMock.On("InsertOnePlayerTransaction", ctx, mock.AnythingOfType("*player.PlayerTransaction")).Return(bson.ObjectID{}, errors.New("transaction failed")).Once()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.AddPlayerMoney(test.ctx, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.PlayerId, result.PlayerId)
				assert.Equal(t, test.expected.Balance, result.Balance)
			}
		})
	}
}

func TestGetPlayerSavingAccount(t *testing.T) {
	repoMock := new(playerRepository.PlayerRepositoryMock)
	usecase := playerUsecase.NewPlayerUsecase(repoMock)

	ctx := context.Background()

	tests := []testGetPlayerSavingAccount{
		{
			name:     "success get player saving account",
			ctx:      ctx,
			playerId: "player:001",
			expected: &player.PlayerSavingAccount{
				PlayerId: "player:001",
				Balance:  250.0,
			},
			isErr: false,
		},
		{
			name:     "failed get player saving account - not found",
			ctx:      ctx,
			playerId: "invalid_player",
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("GetPlayerSavingAccount", ctx, "player:001").Return(&player.PlayerSavingAccount{
		PlayerId: "player:001",
		Balance:  250.0,
	}, nil)

	// Failed case
	repoMock.On("GetPlayerSavingAccount", ctx, "invalid_player").Return(&player.PlayerSavingAccount{}, errors.New("player not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.GetPlayerSavingAccount(test.ctx, test.playerId)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.PlayerId, result.PlayerId)
				assert.Equal(t, test.expected.Balance, result.Balance)
			}
		})
	}
}

func TestFindOnePlayerCredential(t *testing.T) {
	repoMock := new(playerRepository.PlayerRepositoryMock)
	usecase := playerUsecase.NewPlayerUsecase(repoMock)

	ctx := context.Background()
	playerId := bson.NewObjectID()
	testTime := utils.LocalTime()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []testFindOnePlayerCredential{
		{
			name:     "success find player credential",
			ctx:      ctx,
			email:    "test@example.com",
			password: "password123",
			expected: &playerPb.PlayerProfile{
				Id:        playerId.Hex(),
				Email:     "test@example.com",
				Username:  "testuser",
				RoleCode:  0,
				CreatedAt: testTime.String(),
				UpdatedAt: testTime.String(),
			},
			isErr: false,
		},
		{
			name:     "failed find player credential - wrong password",
			ctx:      ctx,
			email:    "test@example.com",
			password: "wrong_password",
			expected: nil,
			isErr:    true,
		},
		{
			name:     "failed find player credential - user not found",
			ctx:      ctx,
			email:    "notfound@example.com",
			password: "password123",
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("FindOnePlayerCredential", ctx, "test@example.com").Return(&player.Player{
		Id:        playerId,
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  string(hashedPassword),
		CreatedAt: testTime,
		UpdatedAt: testTime,
		PlayerRoles: []player.PlayerRole{
			{
				RoleTitle: "Player",
				RoleCode:  0,
			},
		},
	}, nil)

	// Failed case
	repoMock.On("FindOnePlayerCredential", ctx, "notfound@example.com").Return(&player.Player{}, errors.New("user not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.FindOnePlayerCredential(test.ctx, test.email, test.password)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.Id, result.Id)
				assert.Equal(t, test.expected.Email, result.Email)
				assert.Equal(t, test.expected.Username, result.Username)
				assert.Equal(t, test.expected.RoleCode, result.RoleCode)
			}
		})
	}
}
