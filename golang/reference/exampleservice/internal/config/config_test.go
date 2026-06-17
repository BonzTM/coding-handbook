package config

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Env value takes precedence over the hard-coded default; an unset flag
	// falls through to it.
	t.Setenv("HTTP_ADDR", ":9090")
	cfg, err := Load(nil)
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Errorf("Addr = %q, want :9090 (env precedence)", cfg.HTTP.Addr)
	}
	if cfg.HTTP.MaxBodyBytes != defaultMaxBodyBytes {
		t.Errorf("MaxBodyBytes = %d, want default %d", cfg.HTTP.MaxBodyBytes, defaultMaxBodyBytes)
	}
	if cfg.ShutdownGrace != defaultShutdownGrace {
		t.Errorf("ShutdownGrace = %s, want %s", cfg.ShutdownGrace, defaultShutdownGrace)
	}
}

func TestLoadFlagsBeatEnv(t *testing.T) {
	// Precedence: flags > environment. The env sets one value, the flag another.
	t.Setenv("HTTP_ADDR", ":1111")
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := Load([]string{"-http-addr", ":2222"})
	if err != nil {
		t.Fatalf("Load: unexpected error: %v", err)
	}
	if cfg.HTTP.Addr != ":2222" {
		t.Errorf("Addr = %q, want :2222 (flag beats env)", cfg.HTTP.Addr)
	}
	if cfg.Telemetry.LogLevel != slog.LevelWarn {
		t.Errorf("LogLevel = %v, want warn (from env)", cfg.Telemetry.LogLevel)
	}
}

func TestLoadInvalidLogLevel(t *testing.T) {
	t.Setenv("HTTP_ADDR", ":8080")
	_, err := Load([]string{"-log-level", "loud"})
	if err == nil {
		t.Fatal("Load with invalid log level: expected error, got nil")
	}
}

func TestLoadMalformedEnvRejected(t *testing.T) {
	// A malformed env value must be rejected with an actionable, key-named
	// error — never silently defaulted — per the fail-fast contract in
	// golang/foundations/configuration.md.
	tests := []struct {
		name string
		key  string
		bad  string
	}{
		{"malformed int", "DB_MAX_OPEN_CONNS", "abc"},
		{"malformed int64", "HTTP_MAX_BODY_BYTES", "not-a-number"},
		{"malformed duration", "HTTP_READ_TIMEOUT", "15"}, // no unit
		{"malformed bool", "LOG_JSON", "maybe"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HTTP_ADDR", ":8080")
			t.Setenv(tt.key, tt.bad)

			_, err := Load(nil)
			if err == nil {
				t.Fatalf("Load with %s=%q: expected error, got nil (silent default?)", tt.key, tt.bad)
			}
			// The error must name the offending key so the operator can fix it.
			if !strings.Contains(err.Error(), tt.key) {
				t.Errorf("error %q does not name the offending key %q", err, tt.key)
			}
		})
	}
}

func TestLoadMalformedEnvNotMaskedByFlag(t *testing.T) {
	// Even when an explicit flag supplies a valid value, a malformed env default
	// for the SAME setting must still abort: Load checks env parse errors before
	// flag parsing, so the bad operator input is surfaced rather than hidden.
	t.Setenv("HTTP_ADDR", ":8080")
	t.Setenv("DB_MAX_OPEN_CONNS", "abc")

	_, err := Load([]string{"-db-max-open-conns", "10"})
	if err == nil {
		t.Fatal("malformed env masked by valid flag: expected error, got nil")
	}
	if !strings.Contains(err.Error(), "DB_MAX_OPEN_CONNS") {
		t.Errorf("error %q does not name DB_MAX_OPEN_CONNS", err)
	}
}

func TestValidate(t *testing.T) {
	base := func() Config {
		return Config{
			HTTP: HTTPConfig{
				Addr:              ":8080",
				ReadHeaderTimeout: 5 * time.Second,
				MaxBodyBytes:      1 << 20,
			},
			Database: DatabaseConfig{
				MaxOpenConns:    10,
				MaxIdleConns:    10,
				ConnMaxLifetime: time.Minute,
			},
			ShutdownGrace: 10 * time.Second,
		}
	}

	if err := base().Validate(); err != nil {
		t.Fatalf("base config should be valid: %v", err)
	}

	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"empty addr", func(c *Config) { c.HTTP.Addr = "" }},
		{"zero read-header timeout", func(c *Config) { c.HTTP.ReadHeaderTimeout = 0 }},
		{"zero max body", func(c *Config) { c.HTTP.MaxBodyBytes = 0 }},
		{"zero max open conns", func(c *Config) { c.Database.MaxOpenConns = 0 }},
		{"idle exceeds open", func(c *Config) { c.Database.MaxIdleConns = 999 }},
		{"zero conn lifetime", func(c *Config) { c.Database.ConnMaxLifetime = 0 }},
		{"zero shutdown grace", func(c *Config) { c.ShutdownGrace = 0 }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := base()
			tt.mutate(&c)
			if err := c.Validate(); err == nil {
				t.Errorf("Validate(%s): expected error, got nil", tt.name)
			}
		})
	}
}
