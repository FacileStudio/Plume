package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	portepg "github.com/FacileStudio/porte/pg"

	"gorm.io/gorm"
)

// firstImportRetry is the wait after the first failed space import; it
// doubles up to five minutes while Registre does not answer.
const firstImportRetry = 15 * time.Second

// erasureDeps is what the deletion receiver needs from the rest of the app:
// the kit store that deletes the account and the signing service's audit
// trail freeze.
type erasureDeps struct {
	users  *kit.Store
	freeze func(ctx context.Context, docID int64) error
}

// newSpaces builds the spaces service and, when OIDC is configured, the
// Registre link that owns the spaces and the receiver for the people Registre
// erased. Both pull with the app's own OIDC client credentials.
func newSpaces(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger, erasure erasureDeps) (*spaces.Service, *spaces.Registre, *spaces.Deletions) {
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
	deletions := spaces.NewDeletions(spaces.DeletionsConfig{
		ORM: db, Client: client, Users: erasure.users, Freeze: erasure.freeze,
		Erase: portepg.New(sqlDB).EraseIdentity, UploadDir: appEnv.UploadDir, Logger: appLogger,
	})
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
