package auth

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

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

func (controller *Controller) changePassword(context context.Context, userID string, req *ChangePasswordRequest) error {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return errors.Invalid("current and new password required")
	}
	if len(req.NewPassword) < 12 {
		return errors.Invalid("new password must be at least 12 characters")
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errors.Internal("failed to parse user id", err)
	}
	return controller.service.ChangePassword(context, id, req.CurrentPassword, req.NewPassword)
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
