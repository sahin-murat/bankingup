package database

import (
	"context"
	"database/sql"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config interface {
	GetDatabaseURL() string
}

type Database interface {
	Ping(context.Context) error
}

type database struct {
	gormDB *gorm.DB
	sqlDB  *sql.DB
}

func New(cfg Config) (*database, error) {
	databaseURL := cfg.GetDatabaseURL()

	gormDB, err := gorm.Open(postgres.Open(databaseURL))
	if err != nil {
		return nil, fmt.Errorf("can not create PostgreSQL database connection: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("can not get Postgresql connection: %w", err)
	}

	return &database{
		gormDB: gormDB,
		sqlDB:  sqlDB,
	}, nil
}

func (d *database) Ping(ctx context.Context) error {
	if err := d.sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("can not ping PostgreSQL database: %w", err)
	}

	return nil
}

func (d *database) Close() error {
	if err := d.sqlDB.Close(); err != nil {
		return fmt.Errorf("can not close PostgreSQL database: %w", err)
	}

	return nil
}
