package smtp

import (
	"context"
	"errors"
	"testing"

	"github.com/FacileStudio/Plume/apps/api/internal/urlsafe"
	"github.com/FacileStudio/Plume/apps/api/schemas"
)

var privateHosts = []string{"127.0.0.1", "localhost", "10.0.0.5", "192.168.1.10", "169.254.169.254", "::1"}

func TestSaveConfigRefusesPrivateHostsByDefault(t *testing.T) {
	svc := NewService(nil)
	for _, host := range privateHosts {
		req := &SaveConfigRequest{Host: host, Port: 25, FromEmail: "a@example.com"}
		_, err := svc.saveConfig(context.Background(), 1, req)
		if err == nil || err.Error() != "host must not be on a private network" {
			t.Errorf("saveConfig(%s) = %v, want the private network refusal", host, err)
		}
	}
}

func TestSendEmailRefusesPrivateHostsByDefault(t *testing.T) {
	svc := NewService(nil)
	for _, host := range privateHosts {
		config := &schemas.SmtpConfig{Host: host, Port: 25, FromEmail: "a@example.com"}
		err := svc.sendEmail(context.Background(), config, "b@example.com", "s", "b")
		if !errors.Is(err, urlsafe.ErrPrivateAddress) {
			t.Errorf("sendEmail(%s) = %v, want ErrPrivateAddress", host, err)
		}
	}
}

func TestPrivateHostsAreReachedWhenAllowed(t *testing.T) {
	svc := NewService(nil).AllowPrivateHosts(true)
	for _, host := range privateHosts {
		addr, err := svc.dialAddr(context.Background(), host, 25)
		if err != nil || addr == "" {
			t.Errorf("dialAddr(%s) = %q, %v, want an address", host, addr, err)
		}
	}
	config := &schemas.SmtpConfig{Host: "127.0.0.1", Port: 1, FromEmail: "a@example.com"}
	err := svc.sendEmail(context.Background(), config, "b@example.com", "s", "b")
	if err == nil || errors.Is(err, urlsafe.ErrPrivateAddress) {
		t.Fatalf("sendEmail with the opt-out = %v, want a plain connection error", err)
	}
}
