package authUsecase

import (
	"context"
	"time"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/auth"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authRepository"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
)

type (
	AuthUsecaseService interface {
		Login(pctx context.Context, cfg *config.Config, req *auth.PlayerLoginReq) (*auth.ProfileIntercepter, error)
	}

	authUsecase struct {
		authRepository authRepository.AuthRepositoryService
	}
)

func NewAuthUsecase(authRepository authRepository.AuthRepositoryService) AuthUsecaseService {
	return &authUsecase{authRepository}
}

func (u *authUsecase) Login(pctx context.Context, cfg *config.Config, req *auth.PlayerLoginReq) (*auth.ProfileIntercepter, error) {
	profile, err := u.authRepository.CredentialSearch(pctx, cfg.Grpc.PlayerUrl, &playerPb.CredentialSearchReq{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	profile.Id = "player:" + profile.Id

	accessToken := jwtauth.NewAccessToken(cfg.Jwt.AccessSecretKey, cfg.Jwt.AccessDuration, &jwtauth.Claims{
		PlayerId: profile.Id,
		RoleCode: int(profile.RoleCode),
	}).SignToken()

	refreshToken := jwtauth.NewRefreshToken(cfg.Jwt.RefreshSecretKey, cfg.Jwt.RefreshDuration, &jwtauth.Claims{
		PlayerId: profile.Id,
		RoleCode: int(profile.RoleCode),
	}).SignToken()

	credentialId, err := u.authRepository.InsertOneCredential(pctx, &auth.Credential{
		PlayerId:     profile.Id,
		RoleCode:     int(profile.RoleCode),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		CreatedAt:    utils.LocalTime(),
		UpdatedAt:    utils.LocalTime(),
	})
	if err != nil {
		return nil, err
	}

	loc, _ := time.LoadLocation("Asia/Bangkok")

	return &auth.ProfileIntercepter{
			PlayerProfile: &player.PlayerProfile{
				Id:        profile.Id,
				Email:     profile.Email,
				Username:  profile.Username,
				CreatedAt: utils.ConvertStringToTime(profile.CreatedAt).In(loc),
				UpdatedAt: utils.ConvertStringToTime(profile.UpdatedAt).In(loc),
			},
			Credential: &auth.Credential{
				Id: credentialId,
			},
		},
		nil
}
