package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/database"
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/documents"
	"github.com/FacileStudio/Plume/apps/api/modules/fields"
	"github.com/FacileStudio/Plume/apps/api/modules/reminders"
	"github.com/FacileStudio/Plume/apps/api/modules/signers"
	"github.com/FacileStudio/Plume/apps/api/modules/signing"
	"github.com/FacileStudio/Plume/apps/api/modules/smtp"
	"github.com/FacileStudio/Plume/apps/api/modules/spaces"
	"github.com/FacileStudio/Plume/apps/api/modules/verify"
	"github.com/FacileStudio/Plume/apps/api/modules/webhooks"
	"github.com/FacileStudio/Plume/apps/api/schemas"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
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

	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
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

	if err := os.MkdirAll(filepath.Join(appEnv.UploadDir, "avatars"), 0o755); err != nil {
		appLogger.Error("failed to create avatars directory", slog.Any("error", err))
		return err
	}

	svc := newServices(db, appEnv, appLogger)

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

	router := buildRouter(svc, sqlDB, appEnv, appLogger)

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
	auth      *auth.Service
	smtp      *smtp.Service
	webhooks  *webhooks.Service
	documents *documents.Service
	signers   *signers.Service
	fields    *fields.Service
	signing   *signing.Service
	verify    *verify.Service
	spaces    *spaces.Service
	clients   *clients.Service
	reminders *reminders.Service
}

func newServices(db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) *services {
	smtpService := smtp.NewService(db)
	webhookService := webhooks.NewService(db)
	docService := documents.NewService(db, smtpService, webhookService, appEnv.Domain, appEnv.UploadDir)
	return &services{
		auth:      auth.NewService(db, appLogger),
		smtp:      smtpService,
		webhooks:  webhookService,
		documents: docService,
		signers:   signers.NewService(db, docService, webhookService, smtpService, appEnv.Domain),
		fields:    fields.NewService(db, docService),
		signing:   signing.NewService(db, appEnv.UploadDir, docService),
		verify:    verify.NewService(db, docService),
		spaces:    spaces.NewService(db),
		clients:   clients.NewService(db),
		reminders: reminders.NewService(db, smtpService, webhookService, appEnv.Domain),
	}
}

func apiReference() apiref.Config {
	return apiref.Config{
		Title:       "Plume API",
		Description: "Self-hosted document signing platform. All routes are served under the /api prefix.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{
				auth.Documentation,
				documents.Documentation,
				fields.Documentation,
				signers.Documentation,
				signing.Documentation,
				verify.Documentation,
				spaces.Documentation,
				clients.Documentation,
				smtp.Documentation,
				webhooks.Documentation,
			},
		},
	}
}

func buildRouter(svc *services, sqlDB *sql.DB, appEnv env.Config, appLogger *slog.Logger) chi.Router {
	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
	})
	router.Use(middleware.SecurityHeaders)

	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, apiReference())

	router.Route("/api", func(api chi.Router) {
		avatarFS := http.StripPrefix("/api/files/", http.FileServer(http.Dir(appEnv.UploadDir)))
		api.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
			avatarFS.ServeHTTP(w, r)
		})

		auth.RegisterRoutes(api, svc.auth, appEnv)
		documents.RegisterRoutes(api, svc.documents, svc.auth,
			signers.DocumentRoutes(svc.signers),
			fields.DocumentRoutes(svc.fields),
			signing.DocumentRoutes(svc.signing),
		)
		signers.RegisterRoutes(api, svc.signers, svc.auth,
			reminders.SignerRoutes(svc.reminders),
		)
		webhooks.RegisterRoutes(api, svc.webhooks, svc.auth)
		smtp.RegisterRoutes(api, svc.smtp, svc.auth)
		spaces.RegisterRoutes(api, svc.spaces, svc.auth)
		clients.RegisterRoutes(api, svc.clients, svc.auth)

		verifyLimiter := middleware.NewRateLimiter(30, 10).Handler()
		verify.RegisterRoutes(api, svc.verify, verifyLimiter)
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}
