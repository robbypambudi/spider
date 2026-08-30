package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spider/spider/pkg/auth"
	"github.com/spider/spider/pkg/config"
)

func Bootstrap(pool *pgxpool.Pool, settings *config.Settings) error {
	return BootstrapContext(context.Background(), pool, settings)
}

func BootstrapContext(ctx context.Context, pool *pgxpool.Pool, settings *config.Settings) error {
	users := NewUserRepo(pool)
	policies := NewPolicyRepo(pool)

	existing, err := users.GetByEmail(ctx, settings.BootstrapAdminEmail)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("bootstrap user lookup: %w", err)
	}
	if err == pgx.ErrNoRows || existing == nil {
		displayName := "SPIDER Admin"
		_, err = users.Create(ctx, settings.BootstrapAdminEmail,
			auth.HashPassword(settings.BootstrapAdminPassword), "ADMIN", &displayName)
		if err != nil {
			return fmt.Errorf("bootstrap admin: %w", err)
		}
	}

	if err := policies.EnsureDefault(ctx, settings.DefaultSecurityPolicy, settings.DefaultThreshold,
		settings.Chunker, settings.ChunkSize, settings.ChunkOverlap); err != nil {
		return fmt.Errorf("bootstrap policy: %w", err)
	}
	return nil
}
