package config_test

import (
	"strings"
	"testing"
	"time"

	"github.com/example/exampleworker/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	// Not parallel: Load reads the process environment, which other tests set.
	cfg, err := config.Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Errorf("Addr = %q, want :8080", cfg.HTTP.Addr)
	}
	if cfg.Consumer.MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d, want 5", cfg.Consumer.MaxAttempts)
	}
	if cfg.Topic != "widget.events" {
		t.Errorf("Topic = %q, want widget.events", cfg.Topic)
	}
}

func TestLoadFlagsOverride(t *testing.T) {
	cfg, err := config.Load([]string{
		"-http-addr=:9100",
		"-consumer-max-attempts=3",
		"-consumer-base-backoff=50ms",
		"-consumer-max-backoff=5s",
		"-topic=orders",
	})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.HTTP.Addr != ":9100" {
		t.Errorf("Addr = %q, want :9100", cfg.HTTP.Addr)
	}
	if cfg.Consumer.MaxAttempts != 3 {
		t.Errorf("MaxAttempts = %d, want 3", cfg.Consumer.MaxAttempts)
	}
	if cfg.Topic != "orders" {
		t.Errorf("Topic = %q, want orders", cfg.Topic)
	}
}

func TestLoadEnvMalformed(t *testing.T) {
	t.Setenv("CONSUMER_MAX_ATTEMPTS", "not-a-number")
	_, err := config.Load(nil)
	if err == nil {
		t.Fatal("Load: want error for malformed CONSUMER_MAX_ATTEMPTS, got nil")
	}
	if !strings.Contains(err.Error(), "CONSUMER_MAX_ATTEMPTS") {
		t.Errorf("error %q must name the offending key", err)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	base := func() config.Config {
		return config.Config{
			HTTP:          config.HTTPConfig{Addr: ":8080", ReadHeaderTimeout: time.Second},
			Consumer:      config.ConsumerConfig{MaxAttempts: 3, BaseBackoff: time.Millisecond, MaxBackoff: time.Second},
			Outbox:        config.OutboxConfig{PollInterval: time.Second, BatchSize: 10},
			Telemetry:     config.TelemetryConfig{TraceSampleRatio: 1.0},
			Topic:         "t",
			ShutdownGrace: time.Second,
		}
	}

	tests := []struct {
		name    string
		mutate  func(*config.Config)
		wantErr bool
	}{
		{name: "valid", mutate: func(*config.Config) {}, wantErr: false},
		{name: "empty addr", mutate: func(c *config.Config) { c.HTTP.Addr = "" }, wantErr: true},
		{name: "empty topic", mutate: func(c *config.Config) { c.Topic = "" }, wantErr: true},
		{name: "zero attempts", mutate: func(c *config.Config) { c.Consumer.MaxAttempts = 0 }, wantErr: true},
		{name: "backoff inverted", mutate: func(c *config.Config) { c.Consumer.MaxBackoff = 0 }, wantErr: true},
		{name: "bad sample ratio", mutate: func(c *config.Config) { c.Telemetry.TraceSampleRatio = 2 }, wantErr: true},
		{name: "zero outbox batch", mutate: func(c *config.Config) { c.Outbox.BatchSize = 0 }, wantErr: true},
		{name: "zero grace", mutate: func(c *config.Config) { c.ShutdownGrace = 0 }, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := base()
			tc.mutate(&cfg)
			err := cfg.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
