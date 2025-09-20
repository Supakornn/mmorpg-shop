package inventoryUsecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryRepository"
	"github.com/Supakornn/mmorpg-shop/modules/item"
	itemPb "github.com/Supakornn/mmorpg-shop/modules/item/itemPb"
	"github.com/Supakornn/mmorpg-shop/modules/models"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type (
	InventoryUsecaseService interface {
		FindPlayerItems(pctx context.Context, cfg *config.Config, playerId string, req *inventory.InventorySearchReq) (*models.PaginateRes, error)
	}

	inventoryUsecase struct {
		inventoryRepository inventoryRepository.InventoryRepositoryService
	}
)

func NewInventoryUsecase(inventoryRepository inventoryRepository.InventoryRepositoryService) InventoryUsecaseService {
	return &inventoryUsecase{inventoryRepository}
}

func (u *inventoryUsecase) FindPlayerItems(pctx context.Context, cfg *config.Config, playerId string, req *inventory.InventorySearchReq) (*models.PaginateRes, error) {
	filter := bson.D{}
	opts := make([]options.Lister[options.FindOptions], 0)

	if req.Start != "" {
		filter = append(filter, bson.E{Key: "_id", Value: bson.D{{Key: "$gt", Value: utils.ConvertToObjectId(req.Start)}}})
	}

	filter = append(filter, bson.E{Key: "player_id", Value: playerId})

	opts = append(opts, options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	opts = append(opts, options.Find().SetLimit(int64(req.Limit)))

	inventoryData, err := u.inventoryRepository.FindPlayerItems(pctx, filter, opts...)
	if err != nil {
		return nil, errors.New("error: find player items failed")
	}

	itemData, err := u.inventoryRepository.FindItemsInIds(pctx, cfg.Grpc.ItemUrl, &itemPb.FindItemsInIdsReq{
		Ids: func() []string {
			itemsIds := make([]string, 0)
			for _, data := range inventoryData {
				itemsIds = append(itemsIds, data.ItemId)
			}
			return itemsIds
		}(),
	})

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

	results := make([]*inventory.ItemInInventory, 0)
	for _, data := range inventoryData {
		results = append(results, &inventory.ItemInInventory{
			InventoryId: data.Id.Hex(),
			PlayerId:    data.PlayerId,
			ItemShowCase: &item.ItemShowCase{
				ItemId:   data.ItemId,
				Title:    itemMaps[data.ItemId].Title,
				Price:    itemMaps[data.ItemId].Price,
				ImageUrl: itemMaps[data.ItemId].ImageUrl,
				Damage:   itemMaps[data.ItemId].Damage,
			},
		})
	}

	count, err := u.inventoryRepository.CountPlayerItems(pctx, playerId)
	if err != nil {
		return nil, errors.New("error: count player items failed")
	}

	if len(results) == 0 {
		return &models.PaginateRes{
			Data:  make([]*inventory.ItemInInventory, 0),
			Limit: req.Limit,
			Total: count,
			First: models.FirstPaginate{
				Href: fmt.Sprintf("%s/%s?limit=%d", cfg.Paginate.InventoryNextPageBasedUrl, playerId, req.Limit),
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
			Href: fmt.Sprintf("%s/%s?limit=%d", cfg.Paginate.InventoryNextPageBasedUrl, playerId, req.Limit),
		},
		Next: models.NextPaginate{
			Start: results[len(results)-1].InventoryId,
			Href:  fmt.Sprintf("%s/%s?limit=%d&start=%s", cfg.Paginate.InventoryNextPageBasedUrl, playerId, req.Limit, results[len(results)-1].InventoryId),
		},
	}, nil
}
