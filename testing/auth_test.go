package testing

import (
	"context"
	"errors"
	"testing"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/auth"
	authPb "github.com/Supakornn/mmorpg-shop/modules/auth/authPb"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authRepository"
	"github.com/Supakornn/mmorpg-shop/modules/auth/authUsecase"
	"github.com/Supakornn/mmorpg-shop/modules/player"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/Supakornn/mmorpg-shop/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type (
	testLogin struct {
		name     string
		ctx      context.Context
		cfg      *config.Config
		req      *auth.PlayerLoginReq
		expected *auth.ProfileIntercepter
		isErr    bool
	}

	testRefreshToken struct {
		name     string
		ctx      context.Context
		cfg      *config.Config
		req      *auth.RefreshTokenReq
		expected *auth.ProfileIntercepter
		isErr    bool
	}

	testLogout struct {
		name         string
		ctx          context.Context
		credentialId string
		isErr        bool
	}

	testAccessTokenSearch struct {
		name        string
		ctx         context.Context
		accessToken string
		expected    *authPb.AccessTokenSearchRes
		isErr       bool
	}

	testRolesCount struct {
		name     string
		ctx      context.Context
		expected *authPb.RolesCountRes
		isErr    bool
	}
)

func TestLogin(t *testing.T) {
	repoMock := new(authRepository.AuthRepositoryMock)
	usecase := authUsecase.NewAuthUsecase(repoMock)

	cfg := NewTestConfig()
	ctx := context.Background()

	credentialIdSuccess := bson.NewObjectID()
	testTime := utils.LocalTime()

	tests := []testLogin{
		{
			name: "success login",
			ctx:  ctx,
			cfg:  cfg,
			req:  &auth.PlayerLoginReq{Email: "success@test.com", Password: "test123"},
			expected: &auth.ProfileIntercepter{
				PlayerProfile: &player.PlayerProfile{
					Id:        "player:001",
					Email:     "success@test.com",
					Username:  "success_user",
					CreatedAt: testTime,
					UpdatedAt: testTime,
				},
				Credential: &auth.CredentialRes{
					Id:           credentialIdSuccess.Hex(),
					PlayerId:     "player:001",
					RoleCode:     0,
					AccessToken:  "mock_access_token",
					RefreshToken: "mock_refresh_token",
					CreatedAt:    testTime,
					UpdatedAt:    testTime,
				},
			},
			isErr: false,
		},
		{
			name:     "failed login - invalid credentials",
			ctx:      ctx,
			cfg:      cfg,
			req:      &auth.PlayerLoginReq{Email: "failed@test.com", Password: "wrong_password"},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("CredentialSearch", ctx, cfg.Grpc.PlayerUrl, &playerPb.CredentialSearchReq{
		Email: "success@test.com", Password: "test123"},
	).Return(&playerPb.PlayerProfile{
		Id:        "001",
		Email:     "success@test.com",
		Username:  "success_user",
		RoleCode:  0,
		CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
		UpdatedAt: "0001-01-01 00:00:00 +0000 UTC",
	}, nil)

	repoMock.On("InsertOneCredential", ctx, mock.AnythingOfType("*auth.Credential")).Return(credentialIdSuccess, nil)

	repoMock.On("FindOnePlayerCredential", ctx, credentialIdSuccess.Hex()).Return(&auth.Credential{
		Id:           credentialIdSuccess,
		PlayerId:     "player:001",
		RoleCode:     0,
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}, nil)

	// Failed case
	repoMock.On("CredentialSearch", ctx, cfg.Grpc.PlayerUrl, &playerPb.CredentialSearchReq{
		Email: "failed@test.com", Password: "wrong_password"},
	).Return(&playerPb.PlayerProfile{}, errors.New("error: email or password is incorrect"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.Login(test.ctx, test.cfg, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.PlayerProfile.Id, result.PlayerProfile.Id)
				assert.Equal(t, test.expected.PlayerProfile.Email, result.PlayerProfile.Email)
				assert.Equal(t, test.expected.PlayerProfile.Username, result.PlayerProfile.Username)
				assert.Equal(t, test.expected.Credential.PlayerId, result.Credential.PlayerId)
				assert.Equal(t, test.expected.Credential.RoleCode, result.Credential.RoleCode)
				assert.NotEmpty(t, result.Credential.AccessToken)
				assert.NotEmpty(t, result.Credential.RefreshToken)
			}
		})
	}
}

func TestRefreshToken(t *testing.T) {
	repoMock := new(authRepository.AuthRepositoryMock)
	usecase := authUsecase.NewAuthUsecase(repoMock)

	cfg := NewTestConfig()
	ctx := context.Background()

	credentialId := bson.NewObjectID()
	testTime := utils.LocalTime()

	// Create valid JWT tokens for testing
	validRefreshToken := jwtauth.NewRefreshToken(cfg.Jwt.RefreshSecretKey, cfg.Jwt.RefreshDuration, &jwtauth.Claims{
		PlayerId: "player:001",
		RoleCode: 0,
	}).SignToken()

	tests := []testRefreshToken{
		{
			name: "success refresh token",
			ctx:  ctx,
			cfg:  cfg,
			req: &auth.RefreshTokenReq{
				CredentialId: credentialId.Hex(),
				RefreshToken: validRefreshToken,
			},
			expected: &auth.ProfileIntercepter{
				PlayerProfile: &player.PlayerProfile{
					Id:       "player:001",
					Email:    "test@example.com",
					Username: "test_user",
				},
				Credential: &auth.CredentialRes{
					Id:           credentialId.Hex(),
					PlayerId:     "001",
					RoleCode:     0,
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					CreatedAt:    testTime,
					UpdatedAt:    testTime,
				},
			},
			isErr: false,
		},
		{
			name: "failed refresh token - invalid token",
			ctx:  ctx,
			cfg:  cfg,
			req: &auth.RefreshTokenReq{
				CredentialId: credentialId.Hex(),
				RefreshToken: "invalid.jwt.token",
			},
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("FindOnePlayerProfileToRefresh", ctx, cfg.Grpc.PlayerUrl, &playerPb.FindOnePlayerProfileToRefreshReq{
		PlayerId: "001",
	}).Return(&playerPb.PlayerProfile{
		Id:        "001",
		Email:     "test@example.com",
		Username:  "test_user",
		RoleCode:  0,
		CreatedAt: "0001-01-01 00:00:00 +0000 UTC",
		UpdatedAt: "0001-01-01 00:00:00 +0000 UTC",
	}, nil)

	repoMock.On("AccessToken", cfg, &jwtauth.Claims{
		PlayerId: "001",
		RoleCode: 0,
	}).Return("new_access_token")

	repoMock.On("RefreshToken", cfg, &jwtauth.Claims{
		PlayerId: "001",
		RoleCode: 0,
	}).Return("new_refresh_token")

	repoMock.On("UpdateOnePlayerCredential", ctx, credentialId.Hex(), mock.AnythingOfType("*auth.UpdateRefreshTokenReq")).Return(nil)

	repoMock.On("FindOnePlayerCredential", ctx, credentialId.Hex()).Return(&auth.Credential{
		Id:           credentialId,
		PlayerId:     "001",
		RoleCode:     0,
		AccessToken:  "new_access_token",
		RefreshToken: "new_refresh_token",
		CreatedAt:    testTime,
		UpdatedAt:    testTime,
	}, nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.RefreshToken(test.ctx, test.cfg, test.req)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.PlayerProfile.Id, result.PlayerProfile.Id)
				assert.Equal(t, test.expected.PlayerProfile.Email, result.PlayerProfile.Email)
				assert.Equal(t, test.expected.PlayerProfile.Username, result.PlayerProfile.Username)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	repoMock := new(authRepository.AuthRepositoryMock)
	usecase := authUsecase.NewAuthUsecase(repoMock)

	ctx := context.Background()
	credentialId := bson.NewObjectID().Hex()

	tests := []testLogout{
		{
			name:         "success logout",
			ctx:          ctx,
			credentialId: credentialId,
			isErr:        false,
		},
		{
			name:         "failed logout - credential not found",
			ctx:          ctx,
			credentialId: "invalid_credential_id",
			isErr:        true,
		},
	}

	// Success case
	repoMock.On("DeleteOnePlayerCredential", ctx, credentialId).Return(nil)

	// Failed case
	repoMock.On("DeleteOnePlayerCredential", ctx, "invalid_credential_id").Return(errors.New("credential not found"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := usecase.Logout(test.ctx, test.credentialId)

			if test.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAccessTokenSearch(t *testing.T) {
	repoMock := new(authRepository.AuthRepositoryMock)
	usecase := authUsecase.NewAuthUsecase(repoMock)

	ctx := context.Background()
	credentialId := bson.NewObjectID()

	tests := []testAccessTokenSearch{
		{
			name:        "success access token search",
			ctx:         ctx,
			accessToken: "valid_access_token",
			expected: &authPb.AccessTokenSearchRes{
				IsValid: true,
			},
			isErr: false,
		},
		{
			name:        "failed access token search - token not found",
			ctx:         ctx,
			accessToken: "invalid_access_token",
			expected: &authPb.AccessTokenSearchRes{
				IsValid: false,
			},
			isErr: true,
		},
		{
			name:        "failed access token search - token is nil",
			ctx:         ctx,
			accessToken: "nil_access_token",
			expected: &authPb.AccessTokenSearchRes{
				IsValid: false,
			},
			isErr: true,
		},
	}

	// Success case
	repoMock.On("FindOneAccessToken", ctx, "valid_access_token").Return(&auth.Credential{
		Id:          credentialId,
		PlayerId:    "player:001",
		AccessToken: "valid_access_token",
	}, nil)

	// Failed case - not found
	repoMock.On("FindOneAccessToken", ctx, "invalid_access_token").Return(&auth.Credential{}, errors.New("access token not found"))

	// Nil case
	repoMock.On("FindOneAccessToken", ctx, "nil_access_token").Return((*auth.Credential)(nil), nil)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.AccessTokenSearch(test.ctx, test.accessToken)

			if test.isErr {
				assert.Error(t, err)
				assert.Equal(t, test.expected.IsValid, result.IsValid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected.IsValid, result.IsValid)
			}
		})
	}
}

func TestRolesCount(t *testing.T) {
	repoMock := new(authRepository.AuthRepositoryMock)
	usecase := authUsecase.NewAuthUsecase(repoMock)

	ctx := context.Background()

	tests := []testRolesCount{
		{
			name: "success roles count",
			ctx:  ctx,
			expected: &authPb.RolesCountRes{
				Count: 5,
			},
			isErr: false,
		},
		{
			name:     "failed roles count",
			ctx:      ctx,
			expected: nil,
			isErr:    true,
		},
	}

	// Success case
	repoMock.On("RoleCount", ctx).Return(int64(5), nil).Once()

	// Failed case
	repoMock.On("RoleCount", ctx).Return(int64(0), errors.New("database error")).Once()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := usecase.RolesCount(test.ctx)

			if test.isErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, test.expected.Count, result.Count)
			}
		})
	}
}
