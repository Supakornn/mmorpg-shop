package itemRepository

import (
	"context"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ItemRepositoryMock struct {
	mock.Mock
}

func NewItemRepositoryMock() ItemRepositoryService {
	return &ItemRepositoryMock{}
}

func (m *ItemRepositoryMock) IsUniqueItem(pctx context.Context, title string) bool {
	args := m.Called(pctx, title)
	return args.Bool(0)
}

func (m *ItemRepositoryMock) InsertOneItem(pctx context.Context, req *item.Item) (bson.ObjectID, error) {
	args := m.Called(pctx, req)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *ItemRepositoryMock) FindOneItem(pctx context.Context, itemId string) (*item.Item, error) {
	args := m.Called(pctx, itemId)
	return args.Get(0).(*item.Item), args.Error(1)
}

func (m *ItemRepositoryMock) FindManyItems(pctx context.Context, filter bson.D, opts ...options.Lister[options.FindOptions]) ([]*item.ItemShowCase, error) {
	args := m.Called(pctx, filter, opts)
	return args.Get(0).([]*item.ItemShowCase), args.Error(1)
}

func (m *ItemRepositoryMock) CountItems(pctx context.Context, filter bson.D) (int64, error) {
	args := m.Called(pctx, filter)
	return args.Get(0).(int64), args.Error(1)
}

func (m *ItemRepositoryMock) UpdateOneItem(pctx context.Context, itemId string, req bson.M) error {
	args := m.Called(pctx, itemId, req)
	return args.Error(0)
}

func (m *ItemRepositoryMock) UpdateOneItemUsageStatus(pctx context.Context, itemId string, usageStatus bool) error {
	args := m.Called(pctx, itemId, usageStatus)
	return args.Error(0)
}
