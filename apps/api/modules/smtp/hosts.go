package smtp

import (
	"context"
	stderrors "errors"
	"fmt"
	"net"
	"strconv"

	"github.com/FacileStudio/Plume/apps/api/internal/urlsafe"
	"github.com/FacileStudio/Plume/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// AllowPrivateHosts lets owners point their SMTP config at a host on a
// private network, for self-hosted setups with an internal relay. It returns
// the service so it chains off NewService.
func (s *Service) AllowPrivateHosts(allow bool) *Service {
	s.allowPrivate = allow
	return s
}

// dialAddr returns the address to dial for host and port. Unless private
// hosts are allowed it resolves the host and returns a public ip:port, or an
// error wrapping urlsafe.ErrPrivateAddress.
func (s *Service) dialAddr(ctx context.Context, host string, port int) (string, error) {
	portText := strconv.Itoa(port)
	if s.allowPrivate {
		return net.JoinHostPort(host, portText), nil
	}
	return urlsafe.ResolvePublicAddr(ctx, host, portText)
}

// dial opens the TCP connection to the configured server through dialAddr.
func (s *Service) dial(ctx context.Context, config *schemas.SmtpConfig) (net.Conn, error) {
	addr, err := s.dialAddr(ctx, config.Host, config.Port)
	if err != nil {
		return nil, fmt.Errorf("connect %s:%d: %w", config.Host, config.Port, err)
	}
	dialer := &net.Dialer{Timeout: smtpDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("connect %s: %w", addr, err)
	}
	return conn, nil
}

func (s *Service) validateConfigRequest(ctx context.Context, req *SaveConfigRequest) error {
	if req.Host == "" {
		return errors.Invalid("host is required")
	}
	if req.Port == 0 {
		return errors.Invalid("port is required")
	}
	if req.FromEmail == "" {
		return errors.Invalid("from_email is required")
	}
	if _, err := s.dialAddr(ctx, req.Host, req.Port); stderrors.Is(err, urlsafe.ErrPrivateAddress) {
		return errors.Invalid("host must not be on a private network")
	}
	return nil
}
