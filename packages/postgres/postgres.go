package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/demotask/backend/config"
	"github.com/demotask/backend/packages/logger"

	gormLogger "gorm.io/gorm/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Gorm *gorm.DB
	Sql  *sql.DB
}

func NewPostgres(log *logger.Logger) *DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
		config.Get().Postgres.Host,
		config.Get().Postgres.Port,
		config.Get().Postgres.Username,
		config.Get().Postgres.Password,
		config.Get().Postgres.DBName,
		config.Get().Postgres.SSLMode)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		DSN: dsn,
	}), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		log.Fatalf("NewPostgres : %v", err.Error())
	}

	sqldb, err := gormDB.DB()
	if err != nil {
		log.Fatalf("NewPostgres : %v", err.Error())
	}

	if err := sqldb.Ping(); err != nil {
		log.Fatalf("NewPostgres : %v", err.Error())
	}

	sqldb.SetMaxOpenConns(100)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxIdleTime(300 * time.Second)
	sqldb.SetConnMaxLifetime(time.Duration(300 * time.Second))
	return &DB{
		Sql:  sqldb,
		Gorm: gormDB,
	}
}
