package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service is what is left of Plume's authentication after porte took the
// credential: the profile lookup the rest of the app reads, and a thin wrapper
// over porte/local so the register and login routes keep their response shape.
type Service struct {
	orm        *gorm.DB
	sessions   *session.Manager
	passwords  *local.Kit
	logger     *slog.Logger
	controller *Controller
}

// NewService creates a Service wired to the given database, session
// manager, password kit, and logger.
func NewService(orm *gorm.DB, sessions *session.Manager, passwords *local.Kit, logger *slog.Logger) *Service {
	service := &Service{orm: orm, sessions: sessions, passwords: passwords, logger: logger}
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

// Register creates an account through porte/local and signs it in. The cookie
// is set on the way out and the token comes back in the body, so one call
// serves the browser and anything holding the old {user_id, token} shape.
func (service *Service) Register(ctx context.Context, w http.ResponseWriter, r *http.Request, email, password string) (string, string, error) {
	userID, token, err := service.passwords.Register(ctx, w, r, email, "", password)
	if err != nil {
		return "", "", credentialError(err)
	}
	return strconv.FormatInt(userID, 10), token, nil
}

// Login authenticates email and password through porte/local, setting a
// session cookie and returning the user id and token.
func (service *Service) Login(ctx context.Context, w http.ResponseWriter, r *http.Request, email, password string) (string, string, error) {
	userID, token, err := service.passwords.Login(ctx, w, r, email, password)
	if err != nil {
		return "", "", credentialError(err)
	}
	return strconv.FormatInt(userID, 10), token, nil
}

// SetPassword gives a first password to an account that has none. porte
// refuses with porte.ErrPasswordSet when one is already there; replacing it is
// ChangePassword.
func (service *Service) SetPassword(ctx context.Context, userID int64, password string) error {
	return passwordError(service.passwords.SetPassword(ctx, userID, password))
}

// weakPassword is the one sentence both password paths give for a password
// under porte's floor, which is also the floor the controller checks.
func weakPassword() error {
	return errors.Invalid(fmt.Sprintf("password must be at least %d characters", local.DefaultMinPasswordLength))
}

// passwordError replaces porte's sentinel text on the change-password paths
// with wording this app owns. A sentinel is a package contract, not a user
// interface: "porte: invalid email or password" names a field this form does
// not have, and names a dependency to whoever reads it. Each status porte
// chose is kept, and ErrPasswordSet is passed through because the controller
// reads it to tell a missing field from a conflict.
func passwordError(err error) error {
	switch {
	case err == nil:
		return nil
	case stderrors.Is(err, porte.ErrWrongPassword):
		return errors.Unauthorized("current password is incorrect")
	case stderrors.Is(err, porte.ErrNoPassword):
		return errors.Invalid("this account has no password to change")
	case stderrors.Is(err, porte.ErrWeakPassword):
		return weakPassword()
	}
	return err
}

// credentialError is the same substitution for register and login, where the
// email is a field the person filled in and the wording says so.
func credentialError(err error) error {
	switch {
	case err == nil:
		return nil
	case stderrors.Is(err, porte.ErrWrongPassword):
		return errors.Unauthorized("invalid email or password")
	case stderrors.Is(err, porte.ErrEmailTaken):
		return errors.Conflict("an account with this email already exists")
	case stderrors.Is(err, porte.ErrInvalidEmail):
		return errors.Invalid("a valid email is required")
	case stderrors.Is(err, porte.ErrRegistrationClosed):
		return errors.Forbidden("registration is closed on this instance")
	case stderrors.Is(err, porte.ErrWeakPassword):
		return weakPassword()
	}
	return err
}

// Issue mints a named API token: a porte session with a label and no expiry,
// which is what the separate api_tokens table used to be.
func (service *Service) Issue(ctx context.Context, userID int64, label string) (string, porte.Session, error) {
	return service.sessions.Issue(ctx, userID, label)
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

// ChangePassword is PUT /auth/password: confirm the current password, replace
// it, end the account's other logins and rotate the caller's own session. It
// returns the new session token and how many other logins ended.
//
// It takes the request rather than a context because porte writes the rotated
// cookie itself and reads the caller's session id off r.Context(), so this
// cannot be reached from a method holding a bare context.
func (service *Service) ChangePassword(w http.ResponseWriter, r *http.Request, userID int64, current, next string) (string, int64, error) {
	token, revoked, err := service.passwords.ChangePassword(r.Context(), w, r, userID, current, next)
	if err != nil {
		return "", 0, passwordError(err)
	}
	return token, revoked, nil
}

// UpdateProfile changes the name and address.
//
// Since porte v0.3.0 a password identity is keyed on the account id, so an
// address change is this one row and nothing else. The hand-written
// re-key that used to live here is gone with the key it maintained.
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
