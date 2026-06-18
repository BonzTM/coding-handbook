package grpc_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	grpcapi "github.com/example/examplegrpc/internal/api/grpc"
	"github.com/example/examplegrpc/internal/config"
)

// TestServerTransportCredentialsDisabled covers the local/dev path: with no
// cert/key configured the helper returns nil credentials so main falls back to
// an insecure listener.
func TestServerTransportCredentialsDisabled(t *testing.T) {
	creds, err := grpcapi.ServerTransportCredentials(config.TLSConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds != nil {
		t.Fatalf("expected nil credentials when TLS is unconfigured, got %v", creds)
	}
}

// TestServerTransportCredentialsServerTLS covers the secure path: a valid
// cert/key pair yields usable TLS credentials.
func TestServerTransportCredentialsServerTLS(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := writeSelfSignedCert(t, dir)

	creds, err := grpcapi.ServerTransportCredentials(config.TLSConfig{
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("ServerTransportCredentials: %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil credentials")
	}
	if got := creds.Info().SecurityProtocol; got != "tls" {
		t.Errorf("security protocol = %q, want tls", got)
	}
}

// TestServerTransportCredentialsMutualTLS covers the mTLS path: a client-CA
// bundle alongside the server cert produces credentials (client-cert
// verification is enforced at handshake time, which this unit test does not
// drive; it asserts construction succeeds).
func TestServerTransportCredentialsMutualTLS(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile := writeSelfSignedCert(t, dir)

	creds, err := grpcapi.ServerTransportCredentials(config.TLSConfig{
		CertFile:     certFile,
		KeyFile:      keyFile,
		ClientCAFile: certFile, // reuse the self-signed cert as the client CA
	})
	if err != nil {
		t.Fatalf("ServerTransportCredentials (mTLS): %v", err)
	}
	if creds == nil {
		t.Fatal("expected non-nil mTLS credentials")
	}
}

// TestServerTransportCredentialsBadPath covers fail-fast: a configured but
// missing cert is an error, never a silent downgrade to plaintext.
func TestServerTransportCredentialsBadPath(t *testing.T) {
	_, err := grpcapi.ServerTransportCredentials(config.TLSConfig{
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	if err == nil {
		t.Fatal("expected an error for a missing cert/key, got nil")
	}
}

// writeSelfSignedCert generates an ephemeral self-signed ECDSA certificate and
// writes the cert and key PEM files into dir, returning their paths. It is cheap
// (P-256, single cert) so the secure path is exercised without fixtures.
func writeSelfSignedCert(t *testing.T, dir string) (certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "examplegrpc-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certFile = filepath.Join(dir, "cert.pem")
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if werr := os.WriteFile(certFile, certPEM, 0o600); werr != nil {
		t.Fatalf("write cert: %v", werr)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyFile = filepath.Join(dir, "key.pem")
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if werr := os.WriteFile(keyFile, keyPEM, 0o600); werr != nil {
		t.Fatalf("write key: %v", werr)
	}

	return certFile, keyFile
}
