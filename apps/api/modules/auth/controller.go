package auth

import (
	"context"
	stderrors "errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts HTTP requests to Service calls for the auth module.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) register(w http.ResponseWriter, r *http.Request, req *RegisterRequest) (*AuthResponse, error) {
	context := r.Context()
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	if len(req.Password) < 12 {
		return nil, errors.Invalid("password must be at least 12 characters")
	}

	userID, token, err := controller.service.Register(context, w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}

func (controller *Controller) getMe(context context.Context, userID string) (*ProfileResponse, error) {
	user, err := controller.service.getUserByString(context, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		AvatarURL:            user.Avatar(),
		AvatarSource:         user.AvatarOrigin(),
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (controller *Controller) updateMe(context context.Context, userID string, req *UpdateProfileRequest) (*ProfileResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	name := strings.TrimSpace(req.Name)

	if req.ReminderIntervalDays != nil {
		if *req.ReminderIntervalDays < 0 || *req.ReminderIntervalDays > 30 {
			return nil, errors.Invalid("reminder_interval_days must be between 0 and 30")
		}
	}

	user, err := controller.service.updateProfileByString(context, userID, name, email, req.ReminderIntervalDays)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		AvatarURL:            user.Avatar(),
		AvatarSource:         user.AvatarOrigin(),
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
	}, nil
}

// changePassword routes on whether the body carries the current password: with
// it the password is replaced and the caller's session rotated, without it the
// account is adding a first one. porte refuses the second case for an account
// that already has a password, and that refusal is answered as a missing field
// rather than as porte's conflict, because the caller omitted an argument.
func (controller *Controller) changePassword(w http.ResponseWriter, r *http.Request, userID string, req *ChangePasswordRequest) (*ChangePasswordResponse, error) {
	if req.NewPassword == "" {
		return nil, errors.Invalid("new password required")
	}
	if len(req.NewPassword) < 12 {
		return nil, errors.Invalid("new password must be at least 12 characters")
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}

	if req.CurrentPassword == "" {
		if err := controller.service.SetPassword(r.Context(), id, req.NewPassword); err != nil {
			if stderrors.Is(err, porte.ErrPasswordSet) {
				return nil, errors.Invalid("current_password is required to change an existing password")
			}
			return nil, err
		}
		return &ChangePasswordResponse{Status: "ok"}, nil
	}

	token, _, err := controller.service.ChangePassword(w, r, id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return &ChangePasswordResponse{Status: "ok", Token: token}, nil
}

func (controller *Controller) login(w http.ResponseWriter, r *http.Request, req *LoginRequest) (*AuthResponse, error) {
	context := r.Context()
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}

	userID, token, err := controller.service.Login(context, w, r, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}
