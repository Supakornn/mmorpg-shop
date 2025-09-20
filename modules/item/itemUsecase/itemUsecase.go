package itemUsecase

import (
	"context"
	"errors"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemRepository"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
)

type (
	ItemUsecaseService interface {
		CreateItem(pctx context.Context, req *item.CreateItemReq) (*item.ItemShowCase, error)
		FindOneItem(pctx context.Context, itemId string) (*item.ItemShowCase, error)
	}

	itemUsecase struct {
		itemRepository itemRepository.ItemRepositoryService
	}
)

func NewItemUsecase(itemRepository itemRepository.ItemRepositoryService) ItemUsecaseService {
	return &itemUsecase{itemRepository}
}

func (u *itemUsecase) CreateItem(pctx context.Context, req *item.CreateItemReq) (*item.ItemShowCase, error) {
	if !u.itemRepository.IsUniqueItem(pctx, req.Title) {
		return nil, errors.New("error: item already exists")
	}

	itemId, err := u.itemRepository.InsertOneItem(pctx, &item.Item{
		Title:       req.Title,
		Price:       req.Price,
		Damage:      req.Damage,
		UsageStatus: true,
		ImageUrl:    req.ImageUrl,
		CreatedAt:   utils.LocalTime(),
		UpdatedAt:   utils.LocalTime(),
	})
	if err != nil {
		return nil, errors.New("error: insert one item failed")
	}

	return u.FindOneItem(pctx, itemId.Hex())
}

func (u *itemUsecase) FindOneItem(pctx context.Context, itemId string) (*item.ItemShowCase, error) {
	result, err := u.itemRepository.FindOneItem(pctx, itemId)
	if err != nil {
		return nil, err
	}

	return &item.ItemShowCase{
		ItemId:   "item:" + result.Id.Hex(),
		Title:    result.Title,
		Price:    result.Price,
		ImageUrl: result.ImageUrl,
		Damage:   result.Damage,
	}, nil
}
