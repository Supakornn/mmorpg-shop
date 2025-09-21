package paymentUsecase

import (
	"context"
	"errors"
	"log"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/payment"
	"github.com/Supakornn/mmorpg-shop/modules/payment/paymentRepository"
)

type (
	PaymentUsecaseService interface {
		FindeItemsInIds(pctx context.Context, grpcUrl string, req []*payment.ItemServiceReqDatum) error
		GetOffset(pctx context.Context) (int64, error)
	}

	paymentUsecase struct {
		paymentRepository paymentRepository.PaymentRepositoryService
	}
)

func NewPaymentUsecase(paymentRepository paymentRepository.PaymentRepositoryService) PaymentUsecaseService {
	return &paymentUsecase{paymentRepository}
}

func (u *paymentUsecase) FindeItemsInIds(pctx context.Context, grpcUrl string, req []*payment.ItemServiceReqDatum) error {
	setIds := make(map[string]bool)
	for _, v := range req {
		if !setIds[v.ItemId] {
			setIds[v.ItemId] = true
		}
	}

	itemData, err := u.paymentRepository.FindItemsInIds(pctx, grpcUrl, &itemPb.FindItemsInIdsReq{
		Ids: func() []string {
			itemsIds := make([]string, 0)
			for k := range setIds {
				itemsIds = append(itemsIds, k)
			}
			return itemsIds
		}(),
	})
	if err != nil {
		log.Printf("error: find items in ids failed: %v", err.Error())
		return errors.New("error: find items in ids failed")
	}

	itemMaps := make(map[string]*item.ItemShowCase)
	for _, data := range itemData.Items {
		itemMaps[data.Id] = &item.ItemShowCase{
			ItemId:   data.Id,
			Title:    data.Title,
			Price:    data.Price,
			ImageUrl: data.ImageUrl,
			Damage:   int(data.Damage),
		}
	}

	for i := range req {
		if _, ok := itemMaps[req[i].ItemId]; !ok {
			log.Printf("error: item not found: %v", req[i].ItemId)
			return errors.New("error: item not found")
		}

		req[i].Price = itemMaps[req[i].ItemId].Price
	}

	return nil
}

func (u *paymentUsecase) GetOffset(pctx context.Context) (int64, error) {
	return u.paymentRepository.GetOffset(pctx)
}

func (u *paymentUsecase) UpsertOffset(pctx context.Context, offset int64) error {
	return u.paymentRepository.UpsertOffset(pctx, offset)
}
