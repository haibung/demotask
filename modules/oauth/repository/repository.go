package repository

import (
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"go.uber.org/fx"
)

type (
	IOAuthRepository interface {
	}

	OAuthRepository struct {
		fx.In
		DB     *postgres.DB
		Logger *logger.Logger
	}
)

// NewRepository creates a new OAuth repository
func NewRepository(oauthRepository OAuthRepository) IOAuthRepository {
	return &oauthRepository
}
