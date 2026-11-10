package auth

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/tronc/errors"
)

// Controller adapts HTTP requests to Service calls for the auth module.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) getMe(context context.Context, userID int64) (*ProfileResponse, error) {
	user, err := controller.service.GetUser(context, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		AvatarURL:            user.Avatar(),
		AvatarSource:         user.AvatarOrigin(),
		Locale:               user.Locale,
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (controller *Controller) updateMe(context context.Context, userID int64, req *UpdateProfileRequest) (*ProfileResponse, error) {
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

	user, err := controller.service.UpdateProfile(context, userID, name, email, req.ReminderIntervalDays)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		ID:                   strconv.FormatInt(user.ID, 10),
		Email:                user.Email,
		Name:                 user.Name,
		AvatarURL:            user.Avatar(),
		AvatarSource:         user.AvatarOrigin(),
		Locale:               user.Locale,
		ReminderIntervalDays: user.ReminderIntervalDays,
		CreatedAt:            user.CreatedAt.Format(time.RFC3339),
	}, nil
}
