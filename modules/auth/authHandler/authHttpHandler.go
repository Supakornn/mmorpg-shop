package authHandler

import (
	"context"
	"net/http"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/auth"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authUsecase"
	"github.com/Supakornn/mmorpg-shop/pkg/request"
	"github.com/Supakornn/mmorpg-shop/pkg/response"
	"github.com/labstack/echo/v4"
)

type (
	AuthHttpHandlerService interface {
		Login(c echo.Context) error
	}

	authHttpHandler struct {
		cfg         *config.Config
		authUsecase authUsecase.AuthUsecaseService
	}
)

func NewAuthHttpHandler(cfg *config.Config, authUsecase authUsecase.AuthUsecaseService) AuthHttpHandlerService {
	return &authHttpHandler{cfg, authUsecase}
}

func (h *authHttpHandler) Login(c echo.Context) error {
	ctx := context.Background()

	wrapper := request.ContextWrapper(c)

	req := new(auth.PlayerLoginReq)

	if err := wrapper.Bind(req); err != nil {
		return response.ErrResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.authUsecase.Login(ctx, h.cfg, req)
	if err != nil {
		return response.ErrResponse(c, http.StatusUnauthorized, err.Error())
	}

	return response.SuccessResponse(c, http.StatusOK, res)
}
