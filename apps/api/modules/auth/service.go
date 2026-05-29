package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	stderrors "errors"
	"strconv"
	"strings"
	"time"

	"log/slog"

	"api/internal/authcrypto"
	"api/internal/errors"
	"api/internal/oidcavatar"
	"api/schemas"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	storageDir string
	logger     *slog.Logger
	controller *Controller
}

func NewService(orm *gorm.DB, storageDir string, logger *slog.Logger) *Service {
	service := &Service{orm: orm, storageDir: storageDir, logger: logger}
	service.controller = newController(service)
	return service
}

func (service *Service) registerUser(context context.Context, email string, password string) (userID string, token string, err error) {
	hash, err := authcrypto.HashPassword(password)
	if err != nil {
		return "", "", errors.Invalid("invalid password")
	}

	record := &schemas.User{
		Email:        email,
		PasswordHash: hash,
	}
	if err := service.orm.WithContext(context).Create(record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return "", "", errors.Conflict("email already registered")
		}
		return "", "", errors.Internal("failed to create user", err)
	}

	token, err = authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(context, token, record.ID); err != nil {
		return "", "", err
	}

	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) loginUser(context context.Context, email string, password string) (userID string, token string, err error) {
	var record schemas.User
	err = service.orm.WithContext(context).Where("email = ?", email).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", errors.Unauthorized("invalid credentials")
	}
	if err != nil {
		return "", "", errors.Internal("failed to read user", err)
	}
	if !authcrypto.VerifyPassword(password, record.PasswordHash) {
		return "", "", errors.Unauthorized("invalid credentials")
	}

	token, err = authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(context, token, record.ID); err != nil {
		return "", "", err
	}

	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) insertSession(context context.Context, token string, userID int64) error {
	record := &schemas.Session{
		Token:     hashToken(token),
		UserID:    userID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := service.orm.WithContext(context).Create(record).Error; err != nil {
		return errors.Internal("failed to persist session", err)
	}
	return nil
}

func normalizeBearer(authorization string) string {
	value := strings.TrimSpace(authorization)
	if len(value) >= 7 && strings.EqualFold(value[:7], "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return value
}

func (service *Service) authenticateRequest(context context.Context, authorization string) (string, *Data, error) {
	token := normalizeBearer(authorization)
	if token == "" {
		return "", nil, errors.Unauthorized("missing auth token")
	}

	var out struct {
		UserID    int64
		Email     string
		ExpiresAt time.Time
	}
	err := service.orm.WithContext(context).
		Table("sessions s").
		Select("u.id as user_id, u.email as email, s.expires_at as expires_at").
		Joins("join users u on u.id = s.user_id").
		Where("s.token = ?", hashToken(token)).
		Scan(&out).Error
	if err != nil {
		return "", nil, errors.Internal("failed to validate auth token", err)
	}
	if out.UserID == 0 {
		return "", nil, errors.Unauthorized("invalid auth token")
	}
	if time.Now().After(out.ExpiresAt) {
		return "", nil, errors.Unauthorized("expired auth token")
	}

	return strconv.FormatInt(out.UserID, 10), &Data{Email: out.Email}, nil
}

func (service *Service) Authenticate(context context.Context, authorization string) (string, any, error) {
	return service.authenticateRequest(context, authorization)
}

func (service *Service) upsertOIDCUser(context context.Context, email string, profile oidcavatar.Profile, oauth2Token *oauth2.Token) (userID string, token string, err error) {
	var record schemas.User
	err = service.orm.WithContext(context).Where("email = ?", email).First(&record).Error
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", errors.Internal("failed to look up user", err)
	}

	isNew := stderrors.Is(err, gorm.ErrRecordNotFound)
	if isNew {
		record = schemas.User{Email: email}
		if displayName := profile.DisplayName(); displayName != "" {
			record.Name = displayName
		}
		if err := service.orm.WithContext(context).Create(&record).Error; err != nil {
			return "", "", errors.Internal("failed to create user", err)
		}
	}

	dirty := false

	if !isNew {
		if displayName := profile.DisplayName(); displayName != "" {
			record.Name = displayName
			dirty = true
		}
	}

	needsAvatar := profile.Picture != "" && (profile.Picture != record.OIDCPictureURL || (record.AvatarSource != "upload" && record.AvatarURL == ""))
	if needsAvatar && record.AvatarSource != "upload" {
		if record.AvatarURL != "" {
			oldRelative := strings.TrimPrefix(record.AvatarURL, "/files/")
			oidcavatar.RemoveFile(service.storageDir, oldRelative)
		}
		relativePath, fetchErr := oidcavatar.FetchAvatar(profile.Picture, service.storageDir, record.ID, service.logger)
		if fetchErr != nil {
			service.logger.Warn("failed to fetch OIDC avatar", slog.Any("error", fetchErr), slog.Int64("user_id", record.ID))
		} else {
			record.AvatarURL = "/files/" + relativePath
			record.AvatarSource = "oidc"
			dirty = true
		}
		record.OIDCPictureURL = profile.Picture
	}

	if oauth2Token != nil {
		record.OIDCAccessToken = oauth2Token.AccessToken
		record.OIDCRefreshToken = oauth2Token.RefreshToken
		record.OIDCTokenExpiry = oauth2Token.Expiry
		dirty = true
	}

	if dirty {
		if saveErr := service.orm.WithContext(context).Save(&record).Error; saveErr != nil {
			service.logger.Warn("failed to save OIDC profile updates", slog.Any("error", saveErr), slog.Int64("user_id", record.ID))
		}
	}

	token, err = authcrypto.NewToken()
	if err != nil {
		return "", "", errors.Internal("failed to create session", err)
	}
	if err := service.insertSession(context, token, record.ID); err != nil {
		return "", "", err
	}
	return strconv.FormatInt(record.ID, 10), token, nil
}

func (service *Service) SyncOIDCProfile(ctx context.Context, userID string, provider *gooidc.Provider, oauth2Cfg oauth2.Config) error {
	var record schemas.User
	if err := service.orm.WithContext(ctx).Where("id = ?", userID).First(&record).Error; err != nil {
		return errors.NotFound("user not found")
	}

	if record.OIDCAccessToken == "" {
		return errors.Invalid("no OIDC tokens stored for this user")
	}

	if !record.ProfileSyncedAt.IsZero() && time.Since(record.ProfileSyncedAt) < 5*time.Minute {
		return errors.New("resource_exhausted", "profile was synced recently, try again later", nil)
	}

	token := &oauth2.Token{
		AccessToken:  record.OIDCAccessToken,
		RefreshToken: record.OIDCRefreshToken,
		Expiry:       record.OIDCTokenExpiry,
	}

	tokenSource := oauth2Cfg.TokenSource(ctx, token)
	freshToken, err := tokenSource.Token()
	if err != nil {
		return errors.Internal("failed to refresh OIDC token", err)
	}

	userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(freshToken))
	if err != nil {
		return errors.Internal("failed to fetch UserInfo", err)
	}

	var claims struct {
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		GivenName         string `json:"given_name"`
		FamilyName        string `json:"family_name"`
		Picture           string `json:"picture"`
	}
	if err := userInfo.Claims(&claims); err != nil {
		return errors.Internal("failed to parse UserInfo claims", err)
	}

	profile := oidcavatar.Profile{
		Name:             claims.Name,
		PreferredUsername: claims.PreferredUsername,
		GivenName:        claims.GivenName,
		FamilyName:       claims.FamilyName,
		Picture:          claims.Picture,
	}

	dirty := false

	if displayName := profile.DisplayName(); displayName != "" {
		record.Name = displayName
		dirty = true
	}

	needsAvatar := profile.Picture != "" && (profile.Picture != record.OIDCPictureURL || (record.AvatarSource != "upload" && record.AvatarURL == ""))
	if needsAvatar && record.AvatarSource != "upload" {
		if record.AvatarURL != "" {
			oldRelative := strings.TrimPrefix(record.AvatarURL, "/files/")
			oidcavatar.RemoveFile(service.storageDir, oldRelative)
		}
		relativePath, fetchErr := oidcavatar.FetchAvatar(profile.Picture, service.storageDir, record.ID, service.logger)
		if fetchErr != nil {
			service.logger.Warn("failed to fetch OIDC avatar during sync", slog.Any("error", fetchErr), slog.Int64("user_id", record.ID))
		} else {
			record.AvatarURL = "/files/" + relativePath
			record.AvatarSource = "oidc"
			dirty = true
		}
		record.OIDCPictureURL = profile.Picture
	}

	record.OIDCAccessToken = freshToken.AccessToken
	if freshToken.RefreshToken != "" {
		record.OIDCRefreshToken = freshToken.RefreshToken
	}
	record.OIDCTokenExpiry = freshToken.Expiry
	record.ProfileSyncedAt = time.Now()
	dirty = true

	if dirty {
		if saveErr := service.orm.WithContext(ctx).Save(&record).Error; saveErr != nil {
			return errors.Internal("failed to save synced profile", saveErr)
		}
	}

	return nil
}

func (service *Service) getUser(context context.Context, userID string) (*schemas.User, error) {
	var record schemas.User
	err := service.orm.WithContext(context).Where("id = ?", userID).First(&record).Error
	if err != nil {
		return nil, errors.NotFound("user not found")
	}
	return &record, nil
}

func (service *Service) updateUser(context context.Context, userID string, name string, email string, reminderIntervalDays *int) (*schemas.User, error) {
	var record schemas.User
	if err := service.orm.WithContext(context).Where("id = ?", userID).First(&record).Error; err != nil {
		return nil, errors.NotFound("user not found")
	}

	record.Name = name
	record.Email = email
	if reminderIntervalDays != nil {
		record.ReminderIntervalDays = *reminderIntervalDays
	}
	if err := service.orm.WithContext(context).Save(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("email already in use")
		}
		return nil, errors.Internal("failed to update user", err)
	}
	return &record, nil
}

func (service *Service) changePassword(context context.Context, userID string, currentPassword string, newPassword string) error {
	var record schemas.User
	if err := service.orm.WithContext(context).Where("id = ?", userID).First(&record).Error; err != nil {
		return errors.NotFound("user not found")
	}

	if !authcrypto.VerifyPassword(currentPassword, record.PasswordHash) {
		return errors.Unauthorized("current password is incorrect")
	}

	hash, err := authcrypto.HashPassword(newPassword)
	if err != nil {
		return errors.Invalid("invalid password")
	}

	record.PasswordHash = hash
	if err := service.orm.WithContext(context).Save(&record).Error; err != nil {
		return errors.Internal("failed to update password", err)
	}

	if err := service.orm.WithContext(context).Where("user_id = ?", userID).Delete(&schemas.Session{}).Error; err != nil {
		return errors.Internal("failed to revoke sessions", err)
	}

	return nil
}

func (service *Service) CleanupExpiredSessions(ctx context.Context) {
	service.orm.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&schemas.Session{})
}

func StartSessionCleanup(ctx context.Context, service *Service) {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				service.CleanupExpiredSessions(ctx)
			}
		}
	}()
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
