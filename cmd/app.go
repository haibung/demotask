package cmd

import (
	"context"
	"fmt"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/modules"
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	_redis "github.com/demotask/backend/packages/redis"
	"github.com/demotask/backend/routers"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var app = &cobra.Command{
	Use:   "start",
	Short: "Running service",
	Run: func(cmd *cobra.Command, args []string) {
		fx.New(
			fx.Provide(config.Get),
			fx.Provide(routers.NewRouter),
			fx.Provide(postgres.NewPostgres),
			fx.Provide(_redis.NewRedis),
			fx.Provide(logger.NewLogger),
			modules.AppRepository,
			modules.AppController,
			modules.AppRoute,
			fx.Invoke(registerHooks),
		).Run()
	},
}

func init() {
	rootCmd.AddCommand(app)
}

func registerHooks(lifecycle fx.Lifecycle, echoRoute *routers.Router, psql *postgres.DB, redis *redis.Client, logger *logger.Logger) {
	lifecycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				go echoRoute.Start(fmt.Sprintf(":%d", config.Get().Port))
				return nil
			},
			OnStop: func(ctx context.Context) error {
				if err := echoRoute.Shutdown(ctx); err != nil {
					logger.Fatal(err.Error())
					return err
				}
				defer func() {
					if err := psql.Sql.Close(); err != nil {
						logger.Fatal(err.Error())
					}
					if err := redis.Close(); err != nil {
						logger.Fatal(err.Error())
					}
				}()
				return nil
			},
		},
	)
}
