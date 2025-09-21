package paymentRepository

import (
	"context"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	"github.com/stretchr/testify/mock"
)

type PaymentRepositoryMock struct {
	mock.Mock
}

func NewPaymentRepositoryMock() PaymentRepositoryService {
	return &PaymentRepositoryMock{}
}

func (m *PaymentRepositoryMock) FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error) {
	args := m.Called(pctx, grpcUrl, req)
	return args.Get(0).(*itemPb.FindItemsInIdsRes), args.Error(1)
}

func (m *PaymentRepositoryMock) GetOffset(pctx context.Context) (int64, error) {
	args := m.Called(pctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *PaymentRepositoryMock) UpsertOffset(pctx context.Context, offset int64) error {
	args := m.Called(pctx, offset)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) DockedPlayerMoney(pctx context.Context, cfg *config.Config, req *player.CreatePlayerTransactionReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) AddPlayerItem(pctx context.Context, cfg *config.Config, req *inventory.UpdateInventoryReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) RollbackTransaction(pctx context.Context, cfg *config.Config, req *player.RollbackPlayerTransactionReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) RollbackAddPlayerItem(pctx context.Context, cfg *config.Config, req *inventory.RollbackInventoryReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) RemovePlayerItem(pctx context.Context, cfg *config.Config, req *inventory.UpdateInventoryReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) RollbackRemovePlayerItem(pctx context.Context, cfg *config.Config, req *inventory.RollbackInventoryReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *PaymentRepositoryMock) AddPlayerMoney(pctx context.Context, cfg *config.Config, req *player.CreatePlayerTransactionReq) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}
