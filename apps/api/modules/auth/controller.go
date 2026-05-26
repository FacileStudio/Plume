package auth

import (
	"context"
	"strconv"
	"strings"

	"api/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) register(context context.Context, req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, errors.Invalid("invalid email")
	}
	if len(req.Password) < 12 {
		return nil, errors.Invalid("password must be at least 12 characters")
	}

	userID, token, err := controller.service.registerUser(context, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}

func (controller *Controller) getMe(context context.Context, userID string) (*ProfileResponse, error) {
	user, err := controller.service.getUser(context, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format("2006-01-02T15:04:05Z"),
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

	user, err := controller.service.updateUser(context, userID, name, email, req.ReminderIntervalDays)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}

func (controller *Controller) changePassword(context context.Context, userID string, req *ChangePasswordRequest) error {
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return errors.Invalid("current and new password required")
	}
	if len(req.NewPassword) < 12 {
		return errors.Invalid("new password must be at least 12 characters")
	}
	return controller.service.changePassword(context, userID, req.CurrentPassword, req.NewPassword)
}

func (controller *Controller) login(context context.Context, req *LoginRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || req.Password == "" {
		return nil, errors.Invalid("email and password required")
	}

	userID, token, err := controller.service.loginUser(context, email, req.Password)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{UserID: userID, Token: token}, nil
}
