package authRepository

import (
	"context"

	"github.com/Supakornn/mmorpg-shop/config"
	"github.com/Supakornn/mmorpg-shop/modules/auth"
	playerPb "github.com/Supakornn/mmorpg-shop/modules/player/playerPb"
	"github.com/Supakornn/mmorpg-shop/pkg/jwtauth"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthRepositoryMock struct {
	mock.Mock
}

func NewAuthRepositoryMock() AuthRepositoryService {
	return &AuthRepositoryMock{}
}

func (m *AuthRepositoryMock) InsertOneCredential(pctx context.Context, req *auth.Credential) (bson.ObjectID, error) {
	args := m.Called(pctx, req)
	return args.Get(0).(bson.ObjectID), args.Error(1)
}

func (m *AuthRepositoryMock) CredentialSearch(pctx context.Context, grpcUrl string, req *playerPb.CredentialSearchReq) (*playerPb.PlayerProfile, error) {
	args := m.Called(pctx, grpcUrl, req)
	return args.Get(0).(*playerPb.PlayerProfile), args.Error(1)
}

func (m *AuthRepositoryMock) FindOnePlayerCredential(pctx context.Context, credentialId string) (*auth.Credential, error) {
	args := m.Called(pctx, credentialId)
	return args.Get(0).(*auth.Credential), args.Error(1)
}

func (m *AuthRepositoryMock) FindOnePlayerProfileToRefresh(pctx context.Context, grpcUrl string, req *playerPb.FindOnePlayerProfileToRefreshReq) (*playerPb.PlayerProfile, error) {
	args := m.Called(pctx, grpcUrl, req)
	return args.Get(0).(*playerPb.PlayerProfile), args.Error(1)
}

func (m *AuthRepositoryMock) UpdateOnePlayerCredential(pctx context.Context, credentialId string, req *auth.UpdateRefreshTokenReq) error {
	args := m.Called(pctx, credentialId, req)
	return args.Error(0)
}

func (m *AuthRepositoryMock) DeleteOnePlayerCredential(pctx context.Context, credentialId string) error {
	args := m.Called(pctx, credentialId)
	return args.Error(0)
}

func (m *AuthRepositoryMock) FindOneAccessToken(pctx context.Context, accessToken string) (*auth.Credential, error) {
	args := m.Called(pctx, accessToken)
	return args.Get(0).(*auth.Credential), args.Error(1)
}

func (m *AuthRepositoryMock) RoleCount(pctx context.Context) (int64, error) {
	args := m.Called(pctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *AuthRepositoryMock) AccessToken(cfg *config.Config, claims *jwtauth.Claims) string {
	args := m.Called(cfg, claims)
	return args.Get(0).(string)
}

func (m *AuthRepositoryMock) RefreshToken(cfg *config.Config, claims *jwtauth.Claims) string {
	args := m.Called(cfg, claims)
	return args.Get(0).(string)
}
