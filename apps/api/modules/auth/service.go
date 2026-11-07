package auth

import (
	"context"
	stderrors "errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service is what is left of Plume's authentication after porte took the
// credential: the profile lookup the rest of the app reads and the session
// manager the other modules authenticate through.
type Service struct {
	orm        *gorm.DB
	sessions   *session.Manager
	logger     *slog.Logger
	controller *Controller
}

// NewService creates a Service wired to the given database, session manager
// and logger.
func NewService(orm *gorm.DB, sessions *session.Manager, logger *slog.Logger) *Service {
	service := &Service{orm: orm, sessions: sessions, logger: logger}
	service.controller = newController(service)
	return service
}

// RequireAuth is porte's session middleware, re-exported so the module routers
// keep passing this one service to middleware.RequireAuth.
func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return service.sessions.RequireAuth(next)
}

// IdentityForUser turns the user id porte authenticated into the identity the
// rest of Plume reads. It is no longer where authentication happens.
//
// porte deliberately carries neither the email nor any role: what a role may
// do is the app's business, and the profile lives in the app's table. So the
// address is looked up here, which costs the one query the old join cost.
//
// If the user is not found, the session outlived the user: porte's foreign
// key cascades a delete, so this is a race, and it is still not
// authenticated.
func (service *Service) IdentityForUser(ctx context.Context, userID int64) (string, string, error) {
	var out struct {
		ID    int64
		Email string
	}
	err := service.orm.WithContext(ctx).
		Model(&schemas.User{}).
		Select("id", "email").
		Where("id = ?", userID).
		Scan(&out).Error
	if err != nil {
		return "", "", errors.Internal("failed to load the account", err)
	}
	if out.ID == 0 {
		return "", "", errors.Unauthorized("invalid auth token")
	}
	return strconv.FormatInt(out.ID, 10), out.Email, nil
}

// Issue mints a named API token: a porte session with a label and no expiry,
// which is what the separate api_tokens table used to be.
func (service *Service) Issue(ctx context.Context, userID int64, label string) (string, porte.Session, error) {
	return service.sessions.Issue(ctx, userID, label)
}

// revokeAPIToken ends one named token of the user. A login session carries no
// label and answers 404, so this route cannot sign a browser out.
func (service *Service) revokeAPIToken(ctx context.Context, userID, id int64) error {
	sessions, err := service.sessions.List(ctx, userID)
	if err != nil {
		return errors.Internal("failed to list tokens", err)
	}
	for _, session := range sessions {
		if session.ID != id || session.Label == "" {
			continue
		}
		if err := service.sessions.Revoke(ctx, userID, id); err != nil {
			return errors.Internal("failed to revoke the token", err)
		}
		return nil
	}
	return errors.NotFound("api token not found")
}

// AuthenticateRequest resolves the caller of a route that is not mounted
// behind RequireAuth — the inline-image endpoint, which a browser reaches with
// an <img src> and therefore with a cookie and no header.
func (service *Service) AuthenticateRequest(w http.ResponseWriter, r *http.Request) (int64, error) {
	identity, err := service.sessions.Authenticate(w, r)
	if err != nil {
		return 0, err
	}
	return identity.UserID, nil
}

// Sessions exposes the manager for the modules that list or revoke tokens.
func (service *Service) Sessions() *session.Manager { return service.sessions }

// UpdateProfile changes the name and address.
func (service *Service) UpdateProfile(ctx context.Context, userID int64, name, email string, reminderIntervalDays *int) (*schemas.User, error) {
	var record schemas.User
	if err := service.orm.WithContext(ctx).First(&record, userID).Error; err != nil {
		return nil, errors.NotFound("user not found")
	}
	record.Name = name
	record.Email = email
	if reminderIntervalDays != nil {
		record.ReminderIntervalDays = *reminderIntervalDays
	}
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("email already in use")
		}
		return nil, errors.Internal("failed to update user", err)
	}
	return &record, nil
}

// GetUser is the profile the /auth/me routes render.
func (service *Service) GetUser(ctx context.Context, userID int64) (*schemas.User, error) {
	var record schemas.User
	if err := service.orm.WithContext(ctx).First(&record, userID).Error; err != nil {
		return nil, errors.NotFound("user not found")
	}
	return &record, nil
}

// getUserByString and updateProfileByString adapt the decimal-string user id
// the controllers still carry to the int64 porte resolved. Keeping the string
// at the controller boundary leaves the response shapes untouched.
func (service *Service) getUserByString(ctx context.Context, userID string) (*schemas.User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	return service.GetUser(ctx, id)
}

func (service *Service) updateProfileByString(ctx context.Context, userID, name, email string, reminderIntervalDays *int) (*schemas.User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	return service.UpdateProfile(ctx, id, name, email, reminderIntervalDays)
}

// SweepExpiredSessions is the hourly cleanup this app has always run, now over
// porte's table. It is the manager's Sweep rather than a DELETE here, so the
// rule about which rows are spared — named API tokens, which carry no expiry —
// stays in one place instead of being restated by every caller.
func (service *Service) SweepExpiredSessions(ctx context.Context) {
	deleted, err := service.sessions.Sweep(ctx)
	if err != nil {
		service.logger.Warn("failed to sweep expired sessions", slog.Any("error", err))
		return
	}
	if deleted > 0 {
		service.logger.Info("swept expired sessions", slog.Int64("count", deleted))
	}
}

// StartSessionCleanup runs the sweep hourly until the context is cancelled.
func StartSessionCleanup(ctx context.Context, service *Service) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				service.SweepExpiredSessions(ctx)
			}
		}
	}()
}
