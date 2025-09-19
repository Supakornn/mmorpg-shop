package middlewareUsecase

import (
	"errors"
	"log"
	"net/url"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/middleware/middlewareRepository"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/Supakornn/mmorpg-shop/pkg/rbac"
	"github.com/labstack/echo/v4"
)

type (
	MiddlewareUsecaseService interface {
		JwtAuthorization(c echo.Context, cfg *config.Config, accessToken string) (echo.Context, error)
		RbacAuthorization(c echo.Context, cfg *config.Config, expected []int) (echo.Context, error)
		PlayerIdValidation(c echo.Context) (echo.Context, error)
	}

	middlewareUsecase struct {
		middlewareRepository middlewareRepository.MiddlewareRepositoryService
	}
)

func NewMiddlewareUsecase(middlewareRepository middlewareRepository.MiddlewareRepositoryService) MiddlewareUsecaseService {
	return &middlewareUsecase{middlewareRepository}
}

func (u *middlewareUsecase) JwtAuthorization(c echo.Context, cfg *config.Config, accessToken string) (echo.Context, error) {
	ctx := c.Request().Context()

	claims, err := jwtauth.ParseToken(cfg.Jwt.AccessSecretKey, accessToken)
	if err != nil {
		return nil, err
	}

	if err := u.middlewareRepository.AccessTokenSearch(ctx, cfg.Grpc.AuthUrl, accessToken); err != nil {
		return nil, err
	}

	c.Set("player_id", claims.PlayerId)
	c.Set("role_code", claims.RoleCode)

	return c, nil
}

func (u *middlewareUsecase) RbacAuthorization(c echo.Context, cfg *config.Config, expected []int) (echo.Context, error) {
	ctx := c.Request().Context()

	playerRoleCode := c.Get("role_code").(int)

	rolesCount, err := u.middlewareRepository.RolesCount(ctx, cfg.Grpc.AuthUrl)
	if err != nil {
		return nil, err
	}

	plyaerRoleBinary := rbac.IntToBinary(playerRoleCode, int(rolesCount))

	for i := 0; i < int(rolesCount); i++ {
		if plyaerRoleBinary[i]&expected[i] == 1 {
			return c, nil
		}
	}

	return nil, errors.New("error: permission denied")
}

func (u *middlewareUsecase) PlayerIdValidation(c echo.Context) (echo.Context, error) {
	playerIdReq := c.Param("player_id")
	playerIdToken := c.Get("player_id").(string)

	if playerIdToken == "" {
		log.Printf("error: player id not found")
		return nil, errors.New("error: player id not found")
	}

	decodedPlayerIdReq, err := url.QueryUnescape(playerIdReq)
	if err != nil {
		log.Printf("error: failed to decode player id: %s", err.Error())
		return nil, errors.New("error: invalid player id format")
	}

	if decodedPlayerIdReq != playerIdToken {
		log.Printf("error: player id not match: %s != %s", decodedPlayerIdReq, playerIdToken)
		return nil, errors.New("error: player id not match")
	}

	return c, nil
}
