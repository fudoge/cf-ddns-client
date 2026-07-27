package config

import "testing"

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantTimeout int
		wantMode    string
		wantErr     bool
	}{
		{
			name:        "uses defaults",
			args:        []string{},
			wantTimeout: 2,
			wantMode:    "replace",
		},
		{
			name:        "uses provided flags",
			args:        []string{"--mode", "append", "--timeout", "5"},
			wantTimeout: 5,
			wantMode:    "append",
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

			if got.Mode != tt.wantMode {
				t.Fatalf("Mode = %q, want %q", got.Mode, tt.wantMode)
			}
		})
	}
}
