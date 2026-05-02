package tests

import (
	"testing"

	"github.com/zombocoder/go-freebsd-jails-exporter/internal/logging"
)

func TestLoggingNew(t *testing.T) {
	tests := []struct {
		level, format string
		wantErr       bool
	}{
		{"debug", "logfmt", false},
		{"info", "logfmt", false},
		{"warn", "json", false},
		{"error", "json", false},
		{"invalid", "logfmt", true},
		{"info", "invalid", true},
	}

	for _, tt := range tests {
		logger, err := logging.New(tt.level, tt.format)
		if tt.wantErr {
			if err == nil {
				t.Errorf("New(%q, %q) expected error", tt.level, tt.format)
			}
		} else {
			if err != nil {
				t.Errorf("New(%q, %q) unexpected error: %v", tt.level, tt.format, err)
			}
			if logger == nil {
				t.Errorf("New(%q, %q) returned nil logger", tt.level, tt.format)
			}
		}
	}
}
