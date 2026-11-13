package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/database"
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/nuage"
	"github.com/FacileStudio/Plume/apps/api/migrations"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/Plume/apps/api/modules/mcp"
	"github.com/FacileStudio/Plume/apps/api/modules/reminders"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/Plume/apps/api/modules/signing"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/Plume/apps/api/modules/verify"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/oidc"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"

	"gorm.io/gorm"
)

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return err
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return err
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	if handled, err := migrate.Command(context.Background(), os.Args, migrations.Config(sqlDB, appLogger)); handled {
		if err != nil {
			appLogger.Error("migrate", slog.Any("error", err))
		}
		return err
	}

	if err := schemas.MigrateWithIssuer(context.Background(), db, appEnv.IssuerForMigration(), appLogger); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return err
	}

	porteParts, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to build authentication", slog.Any("error", err))
		return err
	}

	svc := newServices(db, appEnv, appLogger, porteParts)
	svc.mcp, err = newMCP(context.Background(), appEnv, porteParts, svc, appLogger)
	if err != nil {
		appLogger.Error("failed to build the MCP server", slog.Any("error", err))
		return err
	}

	go func() {
		count, err := svc.documents.BackfillHashes(context.Background())
		if err != nil {
			appLogger.Warn("hash backfill failed", slog.Any("error", err))
			return
		}
		if count > 0 {
			appLogger.Info("hash backfill complete", slog.Int("documents", count))
		}
	}()

	router, err := buildRouter(svc, sqlDB, appEnv, appLogger, porteParts.oidc)
	if err != nil {
		appLogger.Error("failed to build the router", slog.Any("error", err))
		return err
	}

	addr := ":" + strconv.Itoa(appEnv.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	reminders.Start(shutdownSignal, svc.reminders, appLogger)
	auth.StartSessionCleanup(shutdownSignal, svc.auth)
	startRegistreSync(shutdownSignal, svc)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
			return err
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return err
		}
		appLogger.Info("server stopped")
	}

	return nil
}

type services struct {
	auth       *auth.Service
	smtp       *smtp.Service
	webhooks   *webhooks.Service
	documents  *documents.Service
	signers    *signers.Service
	fields     *fields.Service
	signing    *signing.Service
	verify     *verify.Service
	spaces     *spaces.Service
	registre   *spaces.Registre
	deletions  *spaces.Deletions
	clients    *clients.Service
	reminders  *reminders.Service
	identities kit.IdentityLister
	mcp        *mcp.Service
}

// authParts is what buildAuth assembles: the session manager, the OIDC kit
// and the kit store over Plume's users table.
type authParts struct {
	sessions   *session.Manager
	oidc       *oidc.Kit
	users      *kit.Store
	identities porte.IdentityStore
	erased     porte.ErasedSubjects
}

// buildAuth constructs porte: one session manager and the OIDC kit over the
// identity tables. Sign-in goes through Registre only; there is no local
// password kit. Discovery runs here, so an unreachable or half-configured
// issuer fails at boot rather than on somebody's first login.
//
// The CLI login codes live in a table and are consumed with a DELETE ...
// RETURNING, so a replay finds nothing and a restart loses no pending login.
func buildAuth(ctx context.Context, db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (authParts, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return authParts{}, err
	}
	store, err := newPorteStore(sqlDB)
	if err != nil {
		return authParts{}, err
	}
	resealInBackground(ctx, store, appLogger)
	users, err := auth.NewUsers(ctx, sqlDB, appLogger)
	if err != nil {
		return authParts{}, err
	}
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return authParts{}, err
	}
	var erased porte.ErasedSubjects
	if appEnv.OIDC != nil {
		erased = spaces.ErasedLedger(db, appEnv.OIDC.Issuer)
	}
	oidcKit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:      users,
		Erased:     erased,
		Identities: store.Identities(),
		Sessions:   sessions,
		Codes:      store.LoginCodes(),
		Logger:     appLogger,
		ConfigExtra: func() map[string]any {
			if appEnv.OIDC == nil {
				return nil
			}
			return map[string]any{
				"oidc_redirect_url":   appEnv.OIDC.RedirectURL,
				"oidc_issuer":         appEnv.OIDC.Issuer,
				"registre_spaces_url": spaces.RegistreSpacesURL(appEnv.OIDC.Issuer),
			}
		},
	})
	if err != nil {
		return authParts{}, err
	}
	return authParts{sessions: sessions, oidc: oidcKit, users: users, identities: store.Identities(), erased: erased}, nil
}

func newServices(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger, porteParts authParts) *services {
	smtpService := smtp.NewService(db)
	webhookService := webhooks.NewService(db)
	docService := documents.NewService(db, smtpService, webhookService, appEnv.Domain, appEnv.UploadDir)
	signingService := signing.NewService(db, appEnv.UploadDir, docService)
	spaceService, registre, deletions := newSpaces(db, appEnv, appLogger,
		erasureDeps{users: porteParts.users, freeze: signingService.FreezeAuditTrail, erased: porteParts.erased})
	return &services{
		auth:       auth.NewService(db, porteParts.sessions, porteParts.users, appLogger),
		smtp:       smtpService,
		webhooks:   webhookService,
		documents:  docService,
		signers:    signers.NewService(db, docService, webhookService, smtpService, newNuagePublisher(db, appEnv, appLogger), appEnv.Domain),
		fields:     fields.NewService(db, docService),
		signing:    signingService,
		verify:     verify.NewService(db, docService),
		spaces:     spaceService,
		registre:   registre,
		deletions:  deletions,
		clients:    clients.NewService(db),
		reminders:  reminders.NewService(db, smtpService, webhookService, appEnv.Domain),
		identities: porteParts.identities,
	}
}

func apiReference() apiref.Config {
	return apiref.Config{
		Title:       "Plume API",
		Description: "Self-hosted document signing platform. All routes are served under the /api prefix.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{
				auth.Documentation(),
				documents.Documentation(),
				fields.Documentation(),
				signers.Documentation(),
				signing.Documentation(),
				verify.Documentation(),
				spaces.Documentation(),
				clients.Documentation(),
				smtp.Documentation(),
				webhooks.Documentation(),
				mcp.Documentation(),
				versionDocumentation(),
			},
		},
	}
}

// buildRouter assembles the chi.Router serving the API, wiring middleware,
// auth, and every module's routes.
//
// Behind Traefik and Cloudflare, RemoteAddr is only the visitor if both are
// trusted: Traefik replaces the forwarded chain rather than extending it, so
// the visitor survives in Cf-Connecting-Ip alone. TRUSTED_PROXIES=private,cloudflare
// fills all three.
func buildRouter(svc *services, sqlDB *sql.DB, appEnv env.Config, appLogger *slog.Logger, oidcKit *oidc.Kit) (chi.Router, error) {
	clientDir := spa.DirFromEnv()
	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
		Security: securityConfig(clientDir, appLogger),
	})

	health.Mount(router, health.DB(sqlDB))
	registerVersionRoute(router)
	apiref.Mount(router, apiReference())
	if svc.mcp != nil {
		mcp.RegisterRoutes(router, svc.mcp)
	}

	var mountErr error
	requireAuth := svc.auth.RequireAuth
	router.Route("/api", func(api chi.Router) {
		mountErr = errors.Join(auth.RegisterRoutes(api, svc.auth, oidcKit), mountInvitations(api.With(requireAuth), appEnv, svc))
		documents.RegisterRoutes(api, svc.documents, requireAuth,
			signers.DocumentRoutes(svc.signers),
			fields.DocumentRoutes(svc.fields),
			signing.DocumentRoutes(svc.signing),
		)
		signers.RegisterRoutes(api, svc.signers, requireAuth,
			reminders.SignerRoutes(svc.reminders),
		)
		webhooks.RegisterRoutes(api, svc.webhooks, requireAuth)
		smtp.RegisterRoutes(api, svc.smtp, requireAuth)
		spaces.RegisterRoutes(api, svc.spaces, requireAuth)
		spaces.RegisterSyncRoute(api, svc.registre)
		spaces.RegisterDeletionsRoute(api, svc.deletions)
		clients.RegisterRoutes(api, svc.clients, requireAuth)
		if svc.mcp != nil {
			mcp.RegisterPendingRoutes(api, svc.mcp, requireAuth)
		}

		verify.RegisterRoutes(api, svc.verify, verifyLimiter())
	})

	if mountErr != nil {
		return nil, mountErr
	}

	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router, nil
}

// mountInvitations serves GET /api/me/invitations, the signed-in person's
// pending Registre space invitations for the user menu. It needs Registre,
// so it is mounted only when an issuer is configured.
func mountInvitations(router chi.Router, appEnv env.Config, svc *services) error {
	if appEnv.OIDC == nil {
		return nil
	}
	return kit.MountInvitations(router, kit.InvitationsConfig{
		Directory: directory.Client{
			Issuer:       appEnv.OIDC.Issuer,
			ClientID:     appEnv.OIDC.ClientID,
			ClientSecret: appEnv.OIDC.ClientSecret,
		},
		Identities: svc.identities,
	})
}

// verifyLimiter bounds the public verify routes per visitor: the host of
// RemoteAddr as tronc's real-IP middleware rewrote it, at most 30 requests in
// any minute (a burst of 10, then one every 3 seconds), with the suite's
// rate_limited envelope on refusal.
func verifyLimiter() func(http.Handler) http.Handler {
	return troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{
		Limit:   20.0 / 60,
		Burst:   10,
		MaxKeys: 50000,
		OnRefuse: func(w http.ResponseWriter, _ *http.Request) {
			httpjson.WriteError(w, troncerrors.RateLimited("too many requests"))
		},
	})
}

func newNuagePublisher(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) *nuage.Publisher {
	cfg := appEnv.Nuage
	if cfg == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("nuage archiving disabled", slog.Any("error", err))
		return nil
	}
	store, err := newPorteStore(sqlDB)
	if err != nil {
		appLogger.Error("nuage archiving disabled", slog.Any("error", err))
		return nil
	}
	delegation := nuage.NewDelegation(nuage.DelegationConfig{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
		Audience:     cfg.Audience,
	}, store.Identities(), appLogger)
	appLogger.Info("nuage archiving enabled", slog.String("url", cfg.URL))
	return nuage.NewPublisher(nuage.NewClient(cfg.URL), delegation, appLogger)
}

// externalConnect lists the origins the built client calls besides itself:
// Ardoise's public plan catalog, which muse's PricingPlans reads on the
// landing page.
var externalConnect = []string{"https://api-ardoise.facile.studio"}

// securityConfig extends tronc's default security headers with what the built
// client needs: the inline scripts adapter-static writes into its pages, which
// script-src 'self' would otherwise block, and the external origins above. A
// client directory that cannot be hashed is logged and leaves the default
// script policy, which only reports.
func securityConfig(clientDir string, appLogger *slog.Logger) troncmiddleware.SecurityConfig {
	sources := map[string][]string{"connect-src": externalConnect}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			appLogger.Warn("failed to hash the client's inline scripts", slog.Any("error", err))
		} else {
			sources["script-src"] = hashes
		}
	}
	return troncmiddleware.SecurityConfig{CSPSources: sources}
}
