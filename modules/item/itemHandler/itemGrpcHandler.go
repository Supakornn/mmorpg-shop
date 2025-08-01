package itemHandler

import "github.com/Supakornn/mmorpg-shop/modules/item/itemUsecase"

type (
	itemGrpcHandler struct {
		itemUsecase itemUsecase.ItemUsecaseService
	}
)

func NewItemGrpcHandler(itemUsecase itemUsecase.ItemUsecaseService) *itemGrpcHandler {
	return &itemGrpcHandler{itemUsecase}
}
