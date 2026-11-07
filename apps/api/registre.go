package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

// firstImportRetry is the wait after the first failed space import; it
// doubles up to five minutes while Registre does not answer.
const firstImportRetry = 15 * time.Second

// newSpaces builds the spaces service and, when OIDC is configured, the
// Registre link that owns the spaces and the receiver for the people Registre
// erased. Both pull with the app's own OIDC client credentials.
func newSpaces(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (*spaces.Service, *spaces.Registre, *spaces.Deletions) {
	service := spaces.NewService(db)
	if appEnv.OIDC == nil {
		return service, nil, nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("registre link disabled", slog.Any("error", err))
		return service, nil, nil
	}
	client := directory.Client{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
	}
	deletions := spaces.NewDeletions(db, client, portepg.New(sqlDB).EraseIdentity, appEnv.UploadDir, appLogger)
	return service.ManagedBy(appEnv.OIDC.Issuer), spaces.NewRegistre(db, client, appLogger), deletions
}

// startRegistreSync runs the import and then the sync, and polls the deletion
// list, in the background until ctx ends. Plume runs one replica, so this is
// the one deletion poller. Without OIDC there is no Registre and it does
// nothing.
func startRegistreSync(ctx context.Context, svc *services) {
	if svc.registre == nil {
		return
	}
	go svc.registre.Run(ctx, firstImportRetry)
	go svc.deletions.Run(ctx)
}
