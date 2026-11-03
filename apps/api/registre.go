package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/porte/directory"

	"gorm.io/gorm"
)

// firstImportRetry is the wait after the first failed space import; it
// doubles up to five minutes while Registre does not answer.
const firstImportRetry = 15 * time.Second

// newSpaces builds the spaces service and, when OIDC is configured, the
// Registre link that owns the spaces. It pulls with the app's own OIDC client
// credentials.
func newSpaces(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (*spaces.Service, *spaces.Registre) {
	service := spaces.NewService(db)
	if appEnv.OIDC == nil {
		return service, nil
	}
	client := directory.Client{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
	}
	return service.ManagedBy(appEnv.OIDC.Issuer), spaces.NewRegistre(db, client, appLogger)
}

// startRegistreSync runs the import and then the sync in the background until
// ctx ends. Without OIDC there is no Registre and it does nothing.
func startRegistreSync(ctx context.Context, svc *services) {
	if svc.registre == nil {
		return
	}
	go svc.registre.Run(ctx, firstImportRetry)
}
