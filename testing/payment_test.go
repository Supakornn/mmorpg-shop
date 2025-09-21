package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentRepository"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentUsecase"
	"github.com/stretchr/testify/assert"
)

type (
	testPaymentGetOffset struct {
		name     string
		ctx      context.Context
		expected int64
		isErr    bool
	}

	testPaymentUpsertOffset struct {
		name   string
		ctx    context.Context
		offset int64
		isErr  bool
	}

	testFindItemsInIds struct {
		name    string
		ctx     context.Context
		grpcUrl string
		req     []*payment.ItemServiceReqDatum
		isErr   bool
	}
)

func TestPaymentGetOffset(t *testing.T) {
	repoMock := new(paymentRepository.PaymentRepositoryMock)
	usecase := paymentUsecase.NewPaymentUsecase(repoMock)

	ctx := context.Background()

	tests := []testPaymentGetOffset{
		{
			name:     "success get offset",
			ctx:      ctx,
			expected: 500,
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
	repoMock.On("GetOffset", ctx).Return(int64(500), nil).Once()

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

func TestPaymentUpsertOffset(t *testing.T) {
	repoMock := new(paymentRepository.PaymentRepositoryMock)
	usecase := paymentUsecase.NewPaymentUsecase(repoMock)

	ctx := context.Background()

	tests := []testPaymentUpsertOffset{
		{
			name:   "success upsert offset",
			ctx:    ctx,
			offset: 600,
			isErr:  false,
		},
		{
			name:   "failed upsert offset",
			ctx:    ctx,
			offset: 700,
			isErr:  true,
		},
	}

	// Success case
	repoMock.On("UpsertOffset", ctx, int64(600)).Return(nil)

	// Failed case
	repoMock.On("UpsertOffset", ctx, int64(700)).Return(errors.New("upsert offset failed"))

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

// Note: BuyItem และ SellItem methods ซับซ้อนมากเนื่องจากมี async processing
// และ transaction queue ที่ต้อง mock หลายส่วน ซึ่งเหมาะกับ integration test มากกว่า unit test
