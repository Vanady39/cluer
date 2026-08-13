package main

import (
	"context"
	"time"

	"github.com/Vanady39/cluer/platform/config"
	"github.com/Vanady39/cluer/platform/logger"
	"github.com/Vanady39/cluer/platform/pg"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Logger.Default.Level)

	if cfg.PostgresConfig == nil {
		log.Fatal().Msg("postgres config is required for migrations")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	dsn := cfg.PostgresConfig.GetDSN()
	pool, err := pg.Connect(ctx, dsn, log)
	if err != nil {
		log.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Database connection failed")
	}
	pool.Close()

	if err := pg.Migrate(dsn, log); err != nil {
		log.Fatal().Err(err).Str("dsn", cfg.PostgresConfig.SafeDSN()).Msg("Migrations failed")
	}
}
