package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/FacileStudio/Plume/apps/api/internal/secrets"

	portepg "github.com/FacileStudio/porte/pg"
	"gorm.io/gorm"
)

const porteTokenKeyEnv = "PORTE_TOKEN_KEY"

// tokenSealer parses PORTE_TOKEN_KEY. An unset key returns a nil sealer, which
// leaves sealing off everywhere it is handed.
func tokenSealer() (*portepg.TokenSealer, error) {
	sealer, err := portepg.ParseTokenSealer(os.Getenv(porteTokenKeyEnv))
	if err != nil {
		return nil, fmt.Errorf("%s is set but invalid: %w", porteTokenKeyEnv, err)
	}
	return sealer, nil
}

func newPorteStore(sqlDB *sql.DB) (*portepg.Store, error) {
	sealer, err := tokenSealer()
	if err != nil {
		return nil, err
	}
	return portepg.New(sqlDB, portepg.WithTokenSealer(sealer)), nil
}

// convertSecretsInBackground hashes and seals the signing tokens, SMTP
// passwords and webhook secrets written before this release. It runs beside
// the server: until a row is converted the signing lookup still finds it.
func convertSecretsInBackground(ctx context.Context, db *gorm.DB, sealer *portepg.TokenSealer, logger *slog.Logger) {
	go func() {
		if _, err := secrets.Convert(ctx, db, sealer, logger); err != nil && !errors.Is(err, context.Canceled) {
			logger.Warn("secrets conversion failed", slog.Any("error", err))
		}
	}()
}

func resealInBackground(ctx context.Context, store *portepg.Store, logger *slog.Logger) {
	if os.Getenv(porteTokenKeyEnv) == "" {
		return
	}
	go func() {
		count, err := store.Identities().ResealIdentities(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Warn("porte token reseal failed", slog.Int("resealed", count), slog.Any("error", err))
			return
		}
		logger.Info("porte token reseal complete", slog.Int("resealed", count))
	}()
}
