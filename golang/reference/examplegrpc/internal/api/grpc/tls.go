package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"

	"github.com/example/examplegrpc/internal/config"
)

// ServerTransportCredentials builds the gRPC server TransportCredentials from
// the TLS configuration. It returns (nil, nil) when TLS is not configured so the
// caller can fall back to an insecure local/dev listener; a non-nil error means
// TLS WAS requested but could not be loaded (bad path, malformed PEM), which is
// fail-fast at startup rather than a silent downgrade to plaintext.
//
// When ClientCAFile is set the returned credentials require and verify a client
// certificate (mutual TLS) — the default posture for internal service-to-service
// traffic unless a mesh terminates TLS for us. Otherwise it is one-way server
// TLS. TLS 1.2 is the enforced minimum.
func ServerTransportCredentials(cfg config.TLSConfig) (credentials.TransportCredentials, error) {
	if !cfg.Enabled() {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load server key pair: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if cfg.ClientCAFile != "" {
		pool, err := loadCertPool(cfg.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("load client CA: %w", err)
		}
		// mTLS: require a client cert and verify it against the configured CA.
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
		tlsCfg.ClientCAs = pool
	}

	return credentials.NewTLS(tlsCfg), nil
}

// loadCertPool reads a PEM bundle into an x509.CertPool, failing if no
// certificate could be parsed (an empty or malformed bundle would otherwise
// silently trust nothing).
func loadCertPool(path string) (*x509.CertPool, error) {
	pem, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("no valid certificates found in PEM bundle")
	}
	return pool, nil
}
