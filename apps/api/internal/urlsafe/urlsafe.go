package urlsafe

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

var privateRanges []*net.IPNet

func init() {
	cidrs := []string{
		"0.0.0.0/8",
		"127.0.0.0/8",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"169.254.0.0/16",
		"100.64.0.0/10",
		"198.18.0.0/15",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"fc00::/7",
		"fe80::/10",
		"::1/128",
	}
	for _, cidr := range cidrs {
		_, block, _ := net.ParseCIDR(cidr)
		privateRanges = append(privateRanges, block)
	}
}

func isPrivateIP(ip net.IP) bool {
	if ip.IsUnspecified() {
		return true
	}
	if ip4 := ip.To4(); ip4 != nil {
		ip = ip4
	}
	for _, block := range privateRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func Validate(ctx context.Context, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed")
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("hostname must not be empty")
	}

	addrs, err := net.DefaultResolver.LookupHost(ctx, hostname)
	if err != nil {
		return fmt.Errorf("cannot resolve hostname")
	}

	for _, addr := range addrs {
		ip := net.ParseIP(addr)
		if ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("connections to private network addresses are not allowed")
		}
	}

	return nil
}

func SafeTransport() *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			addrs, err := net.DefaultResolver.LookupHost(ctx, host)
			if err != nil {
				return nil, err
			}

			dialer := &net.Dialer{Timeout: 10 * time.Second}

			for _, resolved := range addrs {
				ip := net.ParseIP(resolved)
				if ip != nil && isPrivateIP(ip) {
					continue
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(resolved, port))
			}

			return nil, fmt.Errorf("connections to private network addresses are not allowed")
		},
	}
}
