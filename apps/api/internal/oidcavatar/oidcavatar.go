package oidcavatar

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxAvatarSize = 5 << 20

type Profile struct {
	Name             string
	PreferredUsername string
	GivenName        string
	FamilyName       string
	Picture          string
}

func (p Profile) DisplayName() string {
	if p.Name != "" {
		return p.Name
	}
	if p.PreferredUsername != "" {
		return p.PreferredUsername
	}
	full := strings.TrimSpace(p.GivenName + " " + p.FamilyName)
	if full != "" {
		return full
	}
	return ""
}

func FetchAvatar(pictureURL, storageDir string, userID int64, logger *slog.Logger) (string, error) {
	parsed, err := url.Parse(pictureURL)
	if err != nil {
		return "", fmt.Errorf("invalid picture URL: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", fmt.Errorf("picture URL must use HTTPS")
	}

	host := parsed.Hostname()
	ips, err := net.LookupIP(host)
	if err != nil {
		return "", fmt.Errorf("DNS lookup failed for %s: %w", host, err)
	}
	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return "", fmt.Errorf("picture URL resolves to private/loopback address")
		}
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pictureURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("avatar fetch returned status %d", resp.StatusCode)
	}

	ct := resp.Header.Get("Content-Type")
	ext, ok := avatarExtension(ct)
	if !ok {
		return "", fmt.Errorf("unsupported content type: %s", ct)
	}

	avatarDir := filepath.Join(storageDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create avatar directory: %w", err)
	}

	filename := fmt.Sprintf("oidc-%d-%d.%s", userID, time.Now().UnixNano(), ext)
	fullPath := filepath.Join(avatarDir, filename)

	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create avatar file: %w", err)
	}
	defer f.Close()

	limited := io.LimitReader(resp.Body, maxAvatarSize+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write avatar: %w", err)
	}
	if n > maxAvatarSize {
		os.Remove(fullPath)
		return "", fmt.Errorf("avatar exceeds %d bytes", maxAvatarSize)
	}

	relativePath := filepath.Join("avatars", filename)
	logger.Info("fetched OIDC avatar", slog.Int64("user_id", userID), slog.String("path", relativePath))
	return relativePath, nil
}

func RemoveFile(storageDir, relativePath string) {
	if !strings.HasPrefix(relativePath, "avatars/") {
		return
	}
	fullPath := filepath.Join(storageDir, relativePath)
	os.Remove(fullPath)
}

func avatarExtension(ct string) (string, bool) {
	ct = strings.ToLower(strings.TrimSpace(ct))
	if idx := strings.Index(ct, ";"); idx != -1 {
		ct = strings.TrimSpace(ct[:idx])
	}
	switch ct {
	case "image/png":
		return "png", true
	case "image/jpeg":
		return "jpg", true
	case "image/gif":
		return "gif", true
	case "image/webp":
		return "webp", true
	default:
		return "", false
	}
}
