package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/FacileStudio/Plume/apps/api/internal/database"
	documentation "github.com/FacileStudio/Plume/apps/api/internal/documentation"
	"github.com/FacileStudio/Plume/apps/api/internal/env"
	"github.com/FacileStudio/Plume/apps/api/internal/httpjson"
	"github.com/FacileStudio/Plume/apps/api/internal/logger"
	"github.com/FacileStudio/Plume/apps/api/internal/middleware"
	"github.com/FacileStudio/Plume/apps/api/modules/auth"
	"github.com/FacileStudio/Plume/apps/api/modules/clients"
	"github.com/FacileStudio/Plume/apps/api/modules/docs"
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
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func main() {
	appEnv, err := env.Load()
	appLogger := logger.New("info")
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return
	}
	appLogger = logger.New(appEnv.LogLevel)

	if appEnv.JournalURL != "" && appEnv.JournalToken != "" {
		journalClient := journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
		defer journalClient.Close()
		appLogger = slog.New(journal.NewHandler(journalClient, appLogger.Handler()))
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return
	}

	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	if err := os.MkdirAll(filepath.Join(appEnv.UploadDir, "avatars"), 0o755); err != nil {
		appLogger.Error("failed to create avatars directory", slog.Any("error", err))
		return
	}

	authService := auth.NewService(db, appEnv.UploadDir, appLogger)
	smtpService := smtp.NewService(db)
	webhookService := webhooks.NewService(db)
	docService := documents.NewService(db, smtpService, webhookService, appEnv.Domain, appEnv.UploadDir)
	signerService := signers.NewService(db, docService, webhookService, smtpService, appEnv.Domain)
	fieldService := fields.NewService(db, docService)
	signingService := signing.NewService(db, appEnv.UploadDir, docService)
	verifyService := verify.NewService(db, docService)
	spaceService := spaces.NewService(db)
	clientService := clients.NewService(db)
	reminderService := reminders.NewService(db, smtpService, webhookService, appEnv.Domain)

	go func() {
		count, err := docService.BackfillHashes(context.Background())
		if err != nil {
			appLogger.Warn("hash backfill failed", slog.Any("error", err))
			return
		}
		if count > 0 {
			appLogger.Info("hash backfill complete", slog.Int("documents", count))
		}
	}()

	docsRegistry := documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			documents.Documentation,
			signers.Documentation,
			webhooks.Documentation,
		},
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.CORS(appEnv.Domain))
	router.Use(middleware.RequestLogger(appLogger))
	router.Use(chimiddleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, request *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/ready", func(w http.ResponseWriter, request *http.Request) {
		readinessContext, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(readinessContext); err != nil {
			httpjson.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})

	router.Route("/api", func(api chi.Router) {
		docs.RegisterRoutes(api, documentation.OpenAPI(docsRegistry))

		avatarFS := http.StripPrefix("/api/files/", http.FileServer(http.Dir(appEnv.UploadDir)))
		api.Get("/files/*", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=86400, immutable")
			avatarFS.ServeHTTP(w, r)
		})

		auth.RegisterRoutes(api, authService, appEnv)
		documents.RegisterRoutes(api, docService, authService,
			signers.DocumentRoutes(signerService),
			fields.DocumentRoutes(fieldService),
			signing.DocumentRoutes(signingService),
		)
		signers.RegisterRoutes(api, signerService, authService,
			reminders.SignerRoutes(reminderService),
		)
		webhooks.RegisterRoutes(api, webhookService, authService)
		smtp.RegisterRoutes(api, smtpService, authService)
		spaces.RegisterRoutes(api, spaceService, authService)
		clients.RegisterRoutes(api, clientService, authService)

		verifyLimiter := middleware.NewRateLimiter(30, 10).Handler()
		verify.RegisterRoutes(api, verifyService, verifyLimiter)
	})

	addr := ":" + appEnv.Port
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

	reminders.Start(shutdownSignal, reminderService, appLogger)
	auth.StartSessionCleanup(shutdownSignal, authService)

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return
		}
		appLogger.Info("server stopped")
	}
}
