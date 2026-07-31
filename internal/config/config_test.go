package config

import (
	"testing"

	"github.com/cloudflare/cloudflare-go/v7/dns"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantTimeout  int
		wantTTL      dns.TTL
		wantMode     string
		wantEndpoint string
		wantJSONPath string
		wantErr      bool
	}{
		{
			name:         "uses defaults",
			args:         []string{},
			wantTimeout:  2,
			wantTTL:      dns.TTL1,
			wantMode:     "replace",
			wantEndpoint: "https://api.ipify.org",
		},
		{
			name:         "uses provided flags",
			args:         []string{"--mode", "append", "--timeout", "5", "--endpoint", "https://example.com/ip", "--jsonpath", "$.ip", "--ttl", "300"},
			wantTimeout:  5,
			wantTTL:      dns.TTL(300),
			wantMode:     "append",
			wantEndpoint: "https://example.com/ip",
			wantJSONPath: "$.ip",
		},
		{
			name:         "accepts automatic TTL",
			args:         []string{"--ttl", "1"},
			wantTimeout:  2,
			wantTTL:      dns.TTL1,
			wantMode:     "replace",
			wantEndpoint: "https://api.ipify.org",
		},
		{
			name:         "accepts minimum TTL",
			args:         []string{"--ttl", "60"},
			wantTimeout:  2,
			wantTTL:      dns.TTL(60),
			wantMode:     "replace",
			wantEndpoint: "https://api.ipify.org",
		},
		{
			name:         "accepts maximum TTL",
			args:         []string{"--ttl", "86400"},
			wantTimeout:  2,
			wantTTL:      dns.TTL(86400),
			wantMode:     "replace",
			wantEndpoint: "https://api.ipify.org",
		},
		{
			name:    "rejects zero timeout",
			args:    []string{"--timeout", "0"},
			wantErr: true,
		},
		{
			name:    "rejects negative timeout",
			args:    []string{"--timeout", "-1"},
			wantErr: true,
		},
		{
			name:    "rejects unsupported mode",
			args:    []string{"--mode", "invalid"},
			wantErr: true,
		},
		{
			name:    "rejects unknown flag",
			args:    []string{"--unknown"},
			wantErr: true,
		},
		{
			name:    "rejects TTL below minimum",
			args:    []string{"--ttl", "59"},
			wantErr: true,
		},
		{
			name:    "rejects TTL above maximum",
			args:    []string{"--ttl", "86401"},
			wantErr: true,
		},
		{
			name:    "rejects fractional TTL",
			args:    []string{"--ttl", "60.5"},
			wantErr: true,
		},
		{
			name:    "rejects non numeric TTL",
			args:    []string{"--ttl", "invalid"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("parseFlags() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Fatalf("parseFlags() error = %v", err)
			}

			if got.Timeout != tt.wantTimeout {
				t.Fatalf("Timeout = %d, want %d", got.Timeout, tt.wantTimeout)
			}

			if got.TTL != tt.wantTTL {
				t.Fatalf("TTL = %v, want %v", got.TTL, tt.wantTTL)
			}

			if got.Mode != tt.wantMode {
				t.Fatalf("Mode = %q, want %q", got.Mode, tt.wantMode)
			}

			if got.Endpoint != tt.wantEndpoint {
				t.Fatalf("Endpoint = %q, want %q", got.Endpoint, tt.wantEndpoint)
			}

			if got.JSONPath != tt.wantJSONPath {
				t.Fatalf("JSONPath = %q, want %q", got.JSONPath, tt.wantJSONPath)
			}
		})
	}
}
