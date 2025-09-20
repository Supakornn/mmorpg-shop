package itemUsecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemRepository"
	"github.com/Supakornn/mmorpg-shop/modules/models"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	ItemUsecaseService interface {
		CreateItem(pctx context.Context, req *item.CreateItemReq) (*item.ItemShowCase, error)
		FindOneItem(pctx context.Context, itemId string) (*item.ItemShowCase, error)
		FindManyItems(pctx context.Context, req *item.ItemSearchReq, basePaginateUrl string) (*models.PaginateRes, error)
		EditItem(pctx context.Context, itemId string, req *item.ItemUpdateReq) (*item.ItemShowCase, error)
		ToggleItemUsageStatus(pctx context.Context, itemId string) (bool, error)
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

func (u *itemUsecase) FindManyItems(pctx context.Context, req *item.ItemSearchReq, basePaginateUrl string) (*models.PaginateRes, error) {
	findItemsFilter := bson.D{}
	findItemsOpts := make([]options.Lister[options.FindOptions], 0)
	countItemsFilter := bson.D{}

	if req.Start != "" {
		req.Start = strings.TrimPrefix(req.Start, "item:")
		findItemsFilter = append(findItemsFilter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: utils.ConvertToObjectId(req.Start)}}})
	}

	if req.Title != "" {
		findItemsFilter = append(findItemsFilter, bson.E{Key: "title", Value: bson.Regex{Pattern: req.Title, Options: "i"}})
		countItemsFilter = append(countItemsFilter, bson.E{Key: "title", Value: bson.Regex{Pattern: req.Title, Options: "i"}})
	}

	findItemsFilter = append(findItemsFilter, bson.E{Key: "usage_status", Value: true})
	countItemsFilter = append(countItemsFilter, bson.E{Key: "usage_status", Value: true})

	findItemsOpts = append(findItemsOpts, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	findItemsOpts = append(findItemsOpts, options.Find().SetLimit(int64(req.Limit)))

	results, err := u.itemRepository.FindManyItems(pctx, findItemsFilter, findItemsOpts...)
	if err != nil {
		return nil, errors.New("error: find many items failed")
	}

	count, err := u.itemRepository.CountItems(pctx, countItemsFilter)
	if err != nil {
		return nil, errors.New("error: count items failed")
	}

	if len(results) == 0 {
		return &models.PaginateRes{
			Data:  make([]*item.ItemShowCase, 0),
			Limit: req.Limit,
			Total: count,
			First: models.FirstPaginate{
				Href: fmt.Sprintf("%s?limit=%d&title=%s", basePaginateUrl, req.Limit, req.Title),
			},
			Next: models.NextPaginate{
				Start: "",
				Href:  "",
			},
		}, nil
	}

	return &models.PaginateRes{
		Data:  results,
		Limit: req.Limit,
		Total: count,
		First: models.FirstPaginate{
			Href: fmt.Sprintf("%s?limit=%d&title=%s", basePaginateUrl, req.Limit, req.Title),
		},
		Next: models.NextPaginate{
			Start: results[len(results)-1].ItemId,
			Href:  fmt.Sprintf("%s?limit=%d&title=%s&start=%s", basePaginateUrl, req.Limit, req.Title, results[len(results)-1].ItemId),
		},
	}, nil
}

func (u *itemUsecase) EditItem(pctx context.Context, itemId string, req *item.ItemUpdateReq) (*item.ItemShowCase, error) {
	updateReq := bson.M{}

	if req.Title != "" {
		if !u.itemRepository.IsUniqueItem(pctx, req.Title) {
			return nil, errors.New("error: item already exists")
		}

		updateReq["title"] = req.Title
	}

	if req.ImageUrl != "" {
		updateReq["image_url"] = req.ImageUrl
	}

	if req.Damage >= 0 {
		updateReq["damage"] = req.Damage
	}

	if req.Price >= 0 {
		updateReq["price"] = req.Price
	}

	updateReq["updated_at"] = utils.LocalTime()

	if err := u.itemRepository.UpdateOneItem(pctx, itemId, updateReq); err != nil {
		return nil, errors.New("error: update one item failed")
	}

	return u.FindOneItem(pctx, itemId)
}

func (u *itemUsecase) ToggleItemUsageStatus(pctx context.Context, itemId string) (bool, error) {
	result, err := u.itemRepository.FindOneItem(pctx, itemId)
	if err != nil {
		return false, errors.New("error: find one item failed")
	}

	if err := u.itemRepository.UpdateOneItemUsageStatus(pctx, itemId, !result.UsageStatus); err != nil {
		return false, errors.New("error: update one item usage status failed")
	}

	return !result.UsageStatus, nil
}
