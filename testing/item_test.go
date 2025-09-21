package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemRepository"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	testCreateItem struct {
		name     string
		ctx      context.Context
		req      *item.CreateItemReq
		expected *item.ItemShowCase
		isErr    bool
	}

	testFindOneItem struct {
		name     string
		ctx      context.Context
		itemId   string
		expected *item.ItemShowCase
		isErr    bool
	}

	testEditItem struct {
		name     string
		ctx      context.Context
		itemId   string
		req      *item.ItemUpdateReq
		expected *item.ItemShowCase
		isErr    bool
	}

	testToggleItemUsageStatus struct {
		name     string
		ctx      context.Context
		itemId   string
		expected bool
		isErr    bool
	}
)

func TestCreateItem(t *testing.T) {
	repoMock := new(itemRepository.ItemRepositoryMock)
	usecase := itemUsecase.NewItemUsecase(repoMock)

	ctx := context.Background()
	itemId := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testCreateItem{
		{
			name: "success create item",
			ctx:  ctx,
			req: &item.CreateItemReq{
				Title:    "Sword of Legends",
				Price:    150.0,
				ImageUrl: "https://example.com/sword.png",
				Damage:   50,
			},
			expected: &item.ItemShowCase{
				ItemId:   "item:" + itemId.Hex(),
				Title:    "Sword of Legends",
				Price:    150.0,
				ImageUrl: "https://example.com/sword.png",
				Damage:   50,
			},
			isErr: false,
		},
		{
			name: "failed create item - item already exists",
			ctx:  ctx,
			req: &item.CreateItemReq{
				Title:    "Existing Sword",
				Price:    100.0,
				ImageUrl: "https://example.com/existing.png",
				Damage:   30,
			},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("IsUniqueItem", ctx, "Sword of Legends").Return(true)
	repoMock.On("InsertOneItem", ctx, mock.AnythingOfType("*item.Item")).Return(itemId, nil)
	repoMock.On("FindOneItem", ctx, itemId.Hex()).Return(&item.Item{
		Id:          itemId,
		Title:       "Sword of Legends",
		Price:       150.0,
		ImageUrl:    "https://example.com/sword.png",
		Damage:      50,
		UsageStatus: true,
		CreatedAt:   testTime,
		UpdatedAt:   testTime,
	}, nil)

	// Failed case
	repoMock.On("IsUniqueItem", ctx, "Existing Sword").Return(false)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.CreateItem(test.ctx, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.ItemId, result.ItemId)
				assert.Equal(t, test.expected.Title, result.Title)
				assert.Equal(t, test.expected.Price, result.Price)
				assert.Equal(t, test.expected.Damage, result.Damage)
			}
		})
	}
}

func TestFindOneItem(t *testing.T) {
	repoMock := new(itemRepository.ItemRepositoryMock)
	usecase := itemUsecase.NewItemUsecase(repoMock)

	ctx := context.Background()
	itemId := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testFindOneItem{
		{
			name:   "success find one item",
			ctx:    ctx,
			itemId: itemId.Hex(),
			expected: &item.ItemShowCase{
				ItemId:   "item:" + itemId.Hex(),
				Title:    "Magic Staff",
				Price:    200.0,
				ImageUrl: "https://example.com/staff.png",
				Damage:   75,
			},
			isErr: false,
		},
		{
			name:     "failed find one item - not found",
			ctx:      ctx,
			itemId:   "invalid_item_id",
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("FindOneItem", ctx, itemId.Hex()).Return(&item.Item{
		Id:          itemId,
		Title:       "Magic Staff",
		Price:       200.0,
		ImageUrl:    "https://example.com/staff.png",
		Damage:      75,
		UsageStatus: true,
		CreatedAt:   testTime,
		UpdatedAt:   testTime,
	}, nil)

	// Failed case
	repoMock.On("FindOneItem", ctx, "invalid_item_id").Return(&item.Item{}, errors.New("item not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.FindOneItem(test.ctx, test.itemId)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.ItemId, result.ItemId)
				assert.Equal(t, test.expected.Title, result.Title)
				assert.Equal(t, test.expected.Price, result.Price)
				assert.Equal(t, test.expected.Damage, result.Damage)
			}
		})
	}
}

func TestEditItem(t *testing.T) {
	repoMock := new(itemRepository.ItemRepositoryMock)
	usecase := itemUsecase.NewItemUsecase(repoMock)

	ctx := context.Background()
	itemId := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testEditItem{
		{
			name:   "success edit item",
			ctx:    ctx,
			itemId: itemId.Hex(),
			req: &item.ItemUpdateReq{
				Title:    "Updated Sword",
				Price:    180.0,
				ImageUrl: "https://example.com/updated.png",
				Damage:   60,
			},
			expected: &item.ItemShowCase{
				ItemId:   "item:" + itemId.Hex(),
				Title:    "Updated Sword",
				Price:    180.0,
				ImageUrl: "https://example.com/updated.png",
				Damage:   60,
			},
			isErr: false,
		},
		{
			name:   "failed edit item - update failed",
			ctx:    ctx,
			itemId: "invalid_item_id",
			req: &item.ItemUpdateReq{
				Title:    "Failed Update",
				Price:    100.0,
				ImageUrl: "https://example.com/failed.png",
				Damage:   25,
			},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("IsUniqueItem", ctx, "Updated Sword").Return(true)
	repoMock.On("UpdateOneItem", ctx, itemId.Hex(), mock.AnythingOfType("bson.M")).Return(nil)
	repoMock.On("FindOneItem", ctx, itemId.Hex()).Return(&item.Item{
		Id:          itemId,
		Title:       "Updated Sword",
		Price:       180.0,
		ImageUrl:    "https://example.com/updated.png",
		Damage:      60,
		UsageStatus: true,
		CreatedAt:   testTime,
		UpdatedAt:   testTime,
	}, nil)

	// Failed case
	repoMock.On("IsUniqueItem", ctx, "Failed Update").Return(true)
	repoMock.On("UpdateOneItem", ctx, "invalid_item_id", mock.AnythingOfType("bson.M")).Return(errors.New("update failed"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.EditItem(test.ctx, test.itemId, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.ItemId, result.ItemId)
				assert.Equal(t, test.expected.Title, result.Title)
				assert.Equal(t, test.expected.Price, result.Price)
				assert.Equal(t, test.expected.Damage, result.Damage)
			}
		})
	}
}

func TestToggleItemUsageStatus(t *testing.T) {
	repoMock := new(itemRepository.ItemRepositoryMock)
	usecase := itemUsecase.NewItemUsecase(repoMock)

	ctx := context.Background()
	itemId := bson.NewObjectID()

	tests := []testToggleItemUsageStatus{
		{
			name:     "success toggle item usage status",
			ctx:      ctx,
			itemId:   itemId.Hex(),
			expected: false,
			isErr:    false,
		},
		{
			name:     "failed toggle item usage status",
			ctx:      ctx,
			itemId:   "invalid_item_id",
			expected: false,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("FindOneItem", ctx, itemId.Hex()).Return(&item.Item{
		Id:          itemId,
		UsageStatus: true,
	}, nil)
	repoMock.On("UpdateOneItemUsageStatus", ctx, itemId.Hex(), false).Return(nil)

	// Failed case
	repoMock.On("FindOneItem", ctx, "invalid_item_id").Return(&item.Item{}, errors.New("item not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.ToggleItemUsageStatus(test.ctx, test.itemId)

			if test.isErr {
				assert.Error(t, err)
				assert.Equal(t, test.expected, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}
