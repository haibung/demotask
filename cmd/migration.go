package cmd

import (
	"github.com/demotask/backend/packages/logger"
	"github.com/demotask/backend/packages/postgres"
	"github.com/pressly/goose"
	"github.com/spf13/cobra"
)

var migration = &cobra.Command{
	Use:   "migrate",
	Short: "Migration database",
	Run: func(cmd *cobra.Command, args []string) {
		log := logger.NewLogger()
		postgres := postgres.NewPostgres(log)

		if len(args) == 0 {
			log.Error("please insert argument up or down")
			return
		}

		switch args[0] {
		case "up":
			if err := Up(postgres); err != nil {
				log.Error(err)
				return
			}
		case "down":
			if err := Down(postgres); err != nil {
				log.Error(err)
				return
			}
		case "create":

			if len(args) < 2 {
				log.Error("please add column name")
				return
			}

			if err := Create(postgres, args[1]); err != nil {
				log.Error(err)
				return
			}
		default:
			log.Error("migration argument not found")
			return
		}
	},
}

// Up :
func Up(db *postgres.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(db.Sql, "sql"); err != nil {
		return err
	}

	return nil
}

// Down :
func Down(db *postgres.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Down(db.Sql, "sql"); err != nil {
		return err
	}

	return nil
}

// Create :
func Create(db *postgres.DB, columnName string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Create(db.Sql, "sql", columnName, "sql"); err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(migration)
}
