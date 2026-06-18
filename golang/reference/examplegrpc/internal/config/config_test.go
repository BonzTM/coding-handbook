package config_test

import (
	"strings"
	"testing"

	"github.com/example/examplegrpc/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := config.Load(nil)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GRPC.Addr != ":9090" {
		t.Errorf("GRPC.Addr = %q, want :9090", cfg.GRPC.Addr)
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Errorf("HTTP.Addr = %q, want :8080", cfg.HTTP.Addr)
	}
	if cfg.Auth.Enabled {
		t.Error("Auth.Enabled should default to false (local/dev)")
	}
}

func TestLoadFlagsOverride(t *testing.T) {
	cfg, err := config.Load([]string{"-grpc-addr", ":7000", "-log-level", "debug"})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.GRPC.Addr != ":7000" {
		t.Errorf("GRPC.Addr = %q, want :7000", cfg.GRPC.Addr)
	}
}

func TestLoadEnvMalformedFailsFast(t *testing.T) {
	t.Setenv("GRPC_MAX_RECV_MSG_BYTES", "not-a-number")
	_, err := config.Load(nil)
	if err == nil {
		t.Fatal("Load accepted a malformed env value")
	}
	if !strings.Contains(err.Error(), "GRPC_MAX_RECV_MSG_BYTES") {
		t.Errorf("error must name the bad key, got: %v", err)
	}
}

func TestValidateAuthTokenRequired(t *testing.T) {
	_, err := config.Load([]string{"-auth-enabled", "-auth-token", ""})
	if err == nil {
		t.Fatal("Load accepted auth enabled with empty token")
	}
	if !strings.Contains(err.Error(), "AUTH_TOKEN") {
		t.Errorf("error must mention AUTH_TOKEN, got: %v", err)
	}
}

func TestValidateSampleRatioRange(t *testing.T) {
	_, err := config.Load([]string{"-trace-sample-ratio", "1.5"})
	if err == nil {
		t.Fatal("Load accepted out-of-range sample ratio")
	}
}

func TestValidateTLSCertKeyRequiredTogether(t *testing.T) {
	_, err := config.Load([]string{"-grpc-tls-cert-file", "/tmp/cert.pem"})
	if err == nil {
		t.Fatal("Load accepted a cert without a key")
	}
	if !strings.Contains(err.Error(), "GRPC_TLS_KEY_FILE") {
		t.Errorf("error must name the missing key, got: %v", err)
	}
}

func TestValidateClientCARequiresServerTLS(t *testing.T) {
	_, err := config.Load([]string{"-grpc-tls-client-ca-file", "/tmp/ca.pem"})
	if err == nil {
		t.Fatal("Load accepted an mTLS client CA without server TLS")
	}
	if !strings.Contains(err.Error(), "GRPC_TLS_CLIENT_CA_FILE") {
		t.Errorf("error must name the client-CA key, got: %v", err)
	}
}

func TestTLSConfigPosture(t *testing.T) {
	if (config.TLSConfig{}).Enabled() {
		t.Error("empty TLSConfig must not be Enabled")
	}
	server := config.TLSConfig{CertFile: "c", KeyFile: "k"}
	if !server.Enabled() || server.MutualTLS() {
		t.Errorf("server TLS posture wrong: enabled=%v mtls=%v", server.Enabled(), server.MutualTLS())
	}
	mtls := config.TLSConfig{CertFile: "c", KeyFile: "k", ClientCAFile: "ca"}
	if !mtls.MutualTLS() {
		t.Error("cert+key+clientCA must be MutualTLS")
	}
}
