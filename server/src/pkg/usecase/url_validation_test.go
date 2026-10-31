package usecase

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateExternalBaseURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		url       string
		wantError bool
		want      string
	}{
		{
			name: "public https url is accepted",
			url:  "https://203.0.113.10",
			want: "https://203.0.113.10",
		},
		{
			name: "trailing slash is removed",
			url:  "https://203.0.113.20/",
			want: "https://203.0.113.20",
		},
		{
			name: "public ip is accepted",
			url:  "https://1.1.1.1",
			want: "https://1.1.1.1",
		},
		{
			name:      "empty url is rejected",
			url:       "",
			wantError: true,
		},
		{
			name:      "non http scheme is rejected",
			url:       "ftp://example.com",
			wantError: true,
		},
		{
			name:      "credentials are rejected",
			url:       "https://user:password@example.com",
			wantError: true,
		},
		{
			name:      "localhost is rejected",
			url:       "http://localhost:8080",
			wantError: true,
		},
		{
			name:      "loopback ip is rejected",
			url:       "http://127.0.0.1:8080",
			wantError: true,
		},
		{
			name:      "ipv6 loopback is rejected",
			url:       "http://[::1]:8080",
			wantError: true,
		},
		{
			name:      "private ip is rejected",
			url:       "http://10.0.0.1",
			wantError: true,
		},
		{
			name:      "link local metadata endpoint is rejected",
			url:       "http://169.254.169.254/latest/meta-data",
			wantError: true,
		},
		{
			name:      "carrier grade nat ip is rejected",
			url:       "http://100.64.0.1",
			wantError: true,
		},
		{
			name:      "query is rejected",
			url:       "https://example.com?foo=bar",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateExternalBaseURL(ctx, tt.url)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateExternalBaseURL_HostnameResolution(t *testing.T) {
	ctx := context.Background()

	originalLookup := lookupIPAddr
	defer func() { lookupIPAddr = originalLookup }()
	lookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
		switch host {
		case "gitlab.example.com":
			return []net.IPAddr{{IP: net.ParseIP("203.0.113.10")}}, nil
		case "internal.example.com":
			return []net.IPAddr{{IP: net.ParseIP("192.168.1.10")}}, nil
		default:
			return nil, &net.DNSError{Err: "no such host", Name: host}
		}
	}

	got, err := validateExternalBaseURL(ctx, "https://gitlab.example.com")
	assert.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com", got)

	_, err = validateExternalBaseURL(ctx, "https://internal.example.com")
	assert.Error(t, err)

	_, err = validateExternalBaseURL(ctx, "https://unknown.example.com")
	assert.Error(t, err)
}
