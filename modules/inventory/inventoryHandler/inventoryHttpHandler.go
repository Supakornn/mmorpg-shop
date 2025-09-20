package inventoryHandler

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/inventory"
	"github.com/Supakornn/mmorpg-shop/modules/inventory/inventoryUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/request"
	"github.com/Supakornn/mmorpg-shop/pkg/response"
	"github.com/labstack/echo/v4"
)

type (
	InventoryHttpHandlerService interface {
		FindPlayerItems(c echo.Context) error
	}

	inventoryHttpHandler struct {
		cfg              *config.Config
		inventoryUsecase inventoryUsecase.InventoryUsecaseService
	}
)

func NewInventoryHttpHandler(cfg *config.Config, inventoryUsecase inventoryUsecase.InventoryUsecaseService) InventoryHttpHandlerService {
	return &inventoryHttpHandler{cfg, inventoryUsecase}
}

func (h *inventoryHttpHandler) FindPlayerItems(c echo.Context) error {
	ctx := context.Background()

	originalParam := c.Param("player_id")

	playerId, err := url.QueryUnescape(originalParam)
	if err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, "invalid parameter format")
	}

	wrapper := request.ContextWrapper(c)

	req := new(inventory.InventorySearchReq)

	if err := wrapper.Bind(req); err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.inventoryUsecase.FindPlayerItems(ctx, h.cfg, playerId, req)
	if err != nil {
		return response.ErrResponse(c, http.StatusInternalServerError, err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, res)
}
