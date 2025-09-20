package itemHandler

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/item"
	"github.com/Supakornn/mmorpg-shop/modules/item/itemUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/request"
	"github.com/Supakornn/mmorpg-shop/pkg/response"
	"github.com/labstack/echo/v4"
)

type (
	ItemHttpHandlerService interface {
		CreateItem(c echo.Context) error
		FindOneItem(c echo.Context) error
		FindManyItems(c echo.Context) error
		EditItem(c echo.Context) error
	}

	itemHttpHandler struct {
		cfg         *config.Config
		itemUsecase itemUsecase.ItemUsecaseService
	}
)

func NewItemHttpHandler(cfg *config.Config, itemUsecase itemUsecase.ItemUsecaseService) ItemHttpHandlerService {
	return &itemHttpHandler{cfg, itemUsecase}
}

func (h *itemHttpHandler) CreateItem(c echo.Context) error {
	ctx := context.Background()

	wrapper := request.ContextWrapper(c)

	req := new(item.CreateItemReq)

	if err := wrapper.Bind(req); err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.itemUsecase.CreateItem(ctx, req)
	if err != nil {
		return response.ErrResponse(c, http.StatusInternalServerError, err.Error())
	}

	return response.SuccessResponse(c, http.StatusCreated, res)
}

func (h *itemHttpHandler) FindOneItem(c echo.Context) error {
	ctx := context.Background()

	originalParam := c.Param("item_id")

	decodedParam, err := url.QueryUnescape(originalParam)
	if err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, "invalid parameter format")
	}

	itemId := strings.TrimPrefix(decodedParam, "item:")

	res, err := h.itemUsecase.FindOneItem(ctx, itemId)
	if err != nil {
		return response.ErrResponse(c, http.StatusInternalServerError, err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, res)
}

func (h *itemHttpHandler) FindManyItems(c echo.Context) error {
	ctx := context.Background()

	wrapper := request.ContextWrapper(c)

	req := new(item.ItemSearchReq)

	if err := wrapper.Bind(req); err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.itemUsecase.FindManyItems(ctx, req, h.cfg.Paginate.ItemNextPageBasedUrl)
	if err != nil {
		return response.ErrResponse(c, http.StatusInternalServerError, err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, res)
}

func (h *itemHttpHandler) EditItem(c echo.Context) error {
	ctx := context.Background()

	originalParam := c.Param("item_id")

	decodedParam, err := url.QueryUnescape(originalParam)
	if err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, "invalid parameter format")
	}

	itemId := strings.TrimPrefix(decodedParam, "item:")

	wrapper := request.ContextWrapper(c)

	req := new(item.ItemUpdateReq)

	if err := wrapper.Bind(req); err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.itemUsecase.EditItem(ctx, itemId, req)
	if err != nil {
		return response.ErrResponse(c, http.StatusInternalServerError, err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, res)
}
