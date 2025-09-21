package inventoryRepository

import (
	"context"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type InventoryRepositoryMock struct {
	mock.Mock
}

func NewInventoryRepositoryMock() InventoryRepositoryService {
	return &InventoryRepositoryMock{}
}

func (m *InventoryRepositoryMock) FindItemsInIds(pctx context.Context, grpcUrl string, req *itemPb.FindItemsInIdsReq) (*itemPb.FindItemsInIdsRes, error) {
	args := m.Called(pctx, grpcUrl, req)
	return args.Get(0).(*itemPb.FindItemsInIdsRes), args.Error(1)
}

func (m *InventoryRepositoryMock) FindPlayerItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*inventory.Inventory, error) {
	args := m.Called(pctx, filter, opts)
	return args.Get(0).([]*inventory.Inventory), args.Error(1)
}

func (m *InventoryRepositoryMock) CountPlayerItems(pctx context.Context, playerId string) (int64, error) {
	args := m.Called(pctx, playerId)
	return args.Get(0).(int64), args.Error(1)
}

func (m *InventoryRepositoryMock) GetOffset(pctx context.Context) (int64, error) {
	args := m.Called(pctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *InventoryRepositoryMock) UpsertOffset(pctx context.Context, offset int64) error {
	args := m.Called(pctx, offset)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) AddPlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) RemovePlayerItemRes(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) RollbackAddPlayerItem(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) RollbackRemovePlayerItem(pctx context.Context, cfg *config.Config, req *payment.PaymentTransferRes) error {
	args := m.Called(pctx, cfg, req)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) InsertOnePlayerItem(pctx context.Context, req *inventory.Inventory) (bson.ObjectID, error) {
	args := m.Called(pctx, req)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *InventoryRepositoryMock) FindOnePlayerItem(pctx context.Context, playerId, itemId string) bool {
	args := m.Called(pctx, playerId, itemId)
	return args.Bool(0)
}

func (m *InventoryRepositoryMock) DeleteOneInventory(pctx context.Context, inventoryId string) error {
	args := m.Called(pctx, inventoryId)
	return args.Error(0)
}

func (m *InventoryRepositoryMock) DeleteOnePlayerItem(pctx context.Context, playerId, itemId string) error {
	args := m.Called(pctx, playerId, itemId)
	return args.Error(0)
}
