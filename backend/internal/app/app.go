package app

import (
	"context"
	"fmt"
	"time"

	commonmongo "decoy-club/backend/internal/common/mongo"
	"decoy-club/backend/internal/config"
	approuter "decoy-club/backend/internal/http"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := commonmongo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		return fmt.Errorf("connect mongo: %w", err)
	}

	router := approuter.NewRouter(&approuter.Dependencies{
		Config:   cfg,
		Database: client.Database(cfg.DatabaseName),
	})

	fmt.Printf("Server is running on port %s\n", cfg.Port)

	if err := router.Run(cfg.Port); err != nil {
		return fmt.Errorf("run server: %w", err)
	}

	return nil
}
