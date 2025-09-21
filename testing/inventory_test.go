package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryRepository"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryUsecase"
	"github.com/stretchr/testify/assert"
)

type (
	testGetOffset struct {
		name     string
		ctx      context.Context
		expected int64
		isErr    bool
	}

	testUpsertOffset struct {
		name   string
		ctx    context.Context
		offset int64
		isErr  bool
	}
)

func TestGetOffset(t *testing.T) {
	repoMock := new(inventoryRepository.InventoryRepositoryMock)
	usecase := inventoryUsecase.NewInventoryUsecase(repoMock)

	ctx := context.Background()

	tests := []testGetOffset{
		{
			name:     "success get offset",
			ctx:      ctx,
			expected: 100,
			isErr:    false,
		},
		{
			name:     "failed get offset",
			ctx:      ctx,
			expected: 0,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("GetOffset", ctx).Return(int64(100), nil).Once()

	// Failed case
	repoMock.On("GetOffset", ctx).Return(int64(0), errors.New("get offset failed")).Once()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.GetOffset(test.ctx)

			if test.isErr {
				assert.Error(t, err)
				assert.Equal(t, int64(0), result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestUpsertOffset(t *testing.T) {
	repoMock := new(inventoryRepository.InventoryRepositoryMock)
	usecase := inventoryUsecase.NewInventoryUsecase(repoMock)

	ctx := context.Background()

	tests := []testUpsertOffset{
		{
			name:   "success upsert offset",
			ctx:    ctx,
			offset: 150,
			isErr:  false,
		},
		{
			name:   "failed upsert offset",
			ctx:    ctx,
			offset: 200,
			isErr:  true,
		},
	}

	// Success case
	repoMock.On("UpsertOffset", ctx, int64(150)).Return(nil)

	// Failed case
	repoMock.On("UpsertOffset", ctx, int64(200)).Return(errors.New("upsert offset failed"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := usecase.UpsertOffset(test.ctx, test.offset)

			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
