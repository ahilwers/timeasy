package usecase

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
)

// validateExternalBaseURL validates a user supplied base URL that is used for
// outbound requests to external services (e.g. self-hosted GitLab or Jira).
//
// Only absolute http/https URLs without embedded credentials are accepted and
// hosts that resolve to loopback, private or link-local addresses are rejected
// to prevent server-side request forgery (SSRF) against internal services.
// The returned URL is normalized (without a trailing slash).
func validateExternalBaseURL(ctx context.Context, rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", fmt.Errorf("base URL must not be empty")
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid base URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("base URL must use http or https")
	}
	if parsed.User != nil {
		return "", fmt.Errorf("base URL must not contain credentials")
	}
	if parsed.Hostname() == "" {
		return "", fmt.Errorf("base URL must contain a host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base URL must not contain a query or fragment")
	}

	if err := validateHost(ctx, parsed.Hostname()); err != nil {
		return "", err
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String(), nil
}

// lookupIPAddr is a variable so tests can resolve host names without network
// access.
var lookupIPAddr = net.DefaultResolver.LookupIPAddr

// validateHost rejects hosts that resolve to addresses which must not be
// reached from the server.
func validateHost(ctx context.Context, host string) error {
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("base URL must not point to localhost")
	}

	var ips []net.IP
	if ip := net.ParseIP(host); ip != nil {
		ips = []net.IP{ip}
	} else {
		resolved, err := lookupIPAddr(ctx, host)
		if err != nil {
			return fmt.Errorf("base URL host could not be resolved: %w", err)
		}
		for _, addr := range resolved {
			ips = append(ips, addr.IP)
		}
	}

	if len(ips) == 0 {
		return fmt.Errorf("base URL host could not be resolved")
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("base URL must not point to an internal address")
		}
	}
	return nil
}

func isBlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() || ip.IsMulticast() {
		return true
	}
	// 100.64.0.0/10 is used for carrier-grade NAT and cloud metadata
	// endpoints in some environments.
	if ip4 := ip.To4(); ip4 != nil {
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return true
		}
	}
	return false
}
