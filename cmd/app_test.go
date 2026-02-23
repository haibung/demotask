package cmd_test

import (
	"testing"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/modules"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/demotask/backend/routers"
	"github.com/go-redis/redis/v8"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// mockConfig Get bypasses the OS env validation for tests
func mockConfig() *config.Config {
	return &config.Config{
		Port: 9999, // use test port
		Auth: struct {
			ExpireAccessToken          string "yaml:\"expireAccessToken\""
			ExpireRefreshToken         string "yaml:\"expireRefreshToken\""
			Secret                     string "yaml:\"secret\""
			SecretClaim                string "yaml:\"secretClaim\""
			ExpireAccessTokenDuration  time.Duration
			ExpireRefreshTokenDuration time.Duration
		}{
			Secret:      "test-secret",
			SecretClaim: "test-claim",
		},
		Redis: struct {
			Host      string "yaml:\"host\""
			Port      string "yaml:\"port\""
			Password  string "yaml:\"password\""
			DefaultDB int    "yaml:\"defaultDb\""
		}{
			Host: "127.0.0.1",
			Port: "6379",
		},
	}
}

// TestAppStartup simulates the exact container dependency injection graph used by the app without starting the server loops
func TestAppStartup(t *testing.T) {
	app := fxtest.New(
		t,
		fx.Provide(mockConfig),
		fx.Provide(routers.NewRouter),
		fx.Provide(logger.NewLogger),

		// Typically you'd provide real or mock DBs here depending on integration strategy
		// fx.Provide(postgres.NewPostgres),
		// fx.Provide(_redis.NewRedis),

		// In a pure DI test, we want to ensure the graph resolves.
		// Mock databases are used to avoid requiring a real postgres instance just to verify DI tree.
		fx.Provide(func() *postgres.DB {
			return &postgres.DB{} // mock uninitialized DB for graph resolution
		}),
		fx.Provide(func() *redis.Client {
			return &redis.Client{} // mock redis
		}),

		modules.AppRepository,
		modules.AppController,
		modules.AppRoute,

		fx.Invoke(func(r *routers.Router) {
			// This func is invoked at the end of the graph resolution.
			// If it executes, the entire tree successfully resolved without cyclic dependencies or missing providers.
			if r == nil {
				t.Fatal("Router dependency was not provided")
			}
		}),
	)

	app.RequireStart()
	app.RequireStop()
}
