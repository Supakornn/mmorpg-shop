package authHandler

import (
	"github.com/Supakornn/mmorpg-shop/modules/auth/authUsecase"
)

type (
	authGrpcHandler struct {
		authUsecase authUsecase.AuthUsecaseService
	}
)

func NewAuthGrpcHandler(authUsecase authUsecase.AuthUsecaseService) authUsecase.AuthUsecaseService {
	return &authGrpcHandler{authUsecase}
}
