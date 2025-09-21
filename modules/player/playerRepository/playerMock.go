package playerRepository

import (
	"context"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PlayerRepositoryMock struct {
	mock.Mock
}

func NewPlayerRepositoryMock() PlayerRepositoryService {
	return &PlayerRepositoryMock{}
}

func (m *PlayerRepositoryMock) IsUniquePlayer(pctx context.Context, email, username string) bool {
	args := m.Called(pctx, email, username)
	return args.Bool(0)
}

func (m *PlayerRepositoryMock) InsertOnePlayer(pctx context.Context, req *player.Player) (bson.ObjectID, error) {
	args := m.Called(pctx, req)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *PlayerRepositoryMock) FindOnePlayerProfile(pctx context.Context, playerId string) (*player.PlayerProfileBson, error) {
	args := m.Called(pctx, playerId)
	return args.Get(0).(*player.PlayerProfileBson), args.Error(1)
}

func (m *PlayerRepositoryMock) InsertOnePlayerTransaction(pctx context.Context, req *player.PlayerTransaction) (bson.ObjectID, error) {
	args := m.Called(pctx, req)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *PlayerRepositoryMock) GetPlayerSavingAccount(pctx context.Context, playerId string) (*player.PlayerSavingAccount, error) {
	args := m.Called(pctx, playerId)
	return args.Get(0).(*player.PlayerSavingAccount), args.Error(1)
}

func (m *PlayerRepositoryMock) FindOnePlayerCredential(pctx context.Context, email string) (*player.Player, error) {
	args := m.Called(pctx, email)
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *PlayerRepositoryMock) FindOnePlayerProfileToRefresh(pctx context.Context, playerId string) (*player.Player, error) {
	args := m.Called(pctx, playerId)
	return args.Get(0).(*player.Player), args.Error(1)
}

func (m *PlayerRepositoryMock) GetOffset(pctx context.Context) (int64, error) {
	args := m.Called(pctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *PlayerRepositoryMock) UpsertOffset(pctx context.Context, offset int64) error {
	args := m.Called(pctx, offset)
	return args.Error(0)
}

func (m *PlayerRepositoryMock) DeleteOnePlayerTransaction(pctx context.Context, transactionId string) error {
	args := m.Called(pctx, transactionId)
	return args.Error(0)
}

func (m *PlayerRepositoryMock) DockedPlayerMoneyRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PlayerRepositoryMock) AddPlayerMoneyRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}
