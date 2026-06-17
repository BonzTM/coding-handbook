// Package config loads, defaults, and validates all process configuration in
// one place. Precedence is flags > environment > hard-coded defaults, per
// golang/foundations/configuration.md. Validation is fail-fast: Load returns a
// fully validated Config or an actionable error, and main aborts before
// opening listeners or external clients.
//
// There are no package-level globals and no init(): everything is wired through
// Load and passed explicitly. Every supported key is documented in .env.example
// and the README.
package config

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

// Config is the closed, required-together set of process settings. It is built
// once by Load and threaded explicitly; it is never read from ambiently.
type Config struct {
	// GRPC holds the gRPC listener and server-hardening settings.
	GRPC GRPCConfig
	// HTTP holds the sidecar listener that serves /metrics, /livez, /readyz.
	HTTP HTTPConfig
	// Telemetry holds logging, tracing, and metrics configuration.
	Telemetry TelemetryConfig
	// Auth gates the bearer-token interceptor.
	Auth AuthConfig
	// ShutdownGrace bounds ordered shutdown. It must exceed worst-case in-flight
	// work and stay under the platform termination grace.
	ShutdownGrace time.Duration
}

// GRPCConfig configures the gRPC server and its self-protection timeouts.
type GRPCConfig struct {
	// Addr is the gRPC listen address, e.g. ":9090".
	Addr string
	// MaxRecvMsgBytes caps inbound message size to bound per-request memory.
	MaxRecvMsgBytes int
	// HandlerTimeout is the deadline-guard ceiling the interceptor imposes when a
	// client sends no deadline, so a single RPC cannot run unbounded.
	HandlerTimeout time.Duration
	// ConnTimeout bounds how long a new connection may take to become ready
	// (handshake + first settings), protecting against slow-loris connections.
	ConnTimeout time.Duration
	// TLS holds optional transport security. When unset the server listens
	// insecure (local/dev only); production must configure it.
	TLS TLSConfig
}

// TLSConfig configures transport security for the gRPC listener. It is
// config-gated: with CertFile and KeyFile set the server presents that
// certificate and serves over TLS; additionally setting ClientCAFile turns on
// mutual TLS (the server requires and verifies a client certificate). Leaving
// CertFile/KeyFile empty selects the insecure local/dev listener.
//
// mTLS is the default posture for internal service-to-service traffic: callers
// present a client cert signed by ClientCAFile unless a service mesh terminates
// TLS for us. The reference defaults to insecure so it boots offline, and logs a
// loud warning that production requires TLS.
type TLSConfig struct {
	// CertFile is the PEM server certificate (chain) path. Empty disables TLS.
	CertFile string
	// KeyFile is the PEM private key path matching CertFile. Empty disables TLS.
	KeyFile string
	// ClientCAFile is the PEM CA bundle used to verify client certificates. When
	// set (alongside CertFile/KeyFile) the server enforces mutual TLS.
	ClientCAFile string
}

// Enabled reports whether server-side TLS is configured (a cert and key are
// both present). The server constructs TLS credentials only when this is true.
func (t TLSConfig) Enabled() bool { return t.CertFile != "" && t.KeyFile != "" }

// MutualTLS reports whether client-certificate verification (mTLS) is enabled,
// which requires both server TLS and a client-CA bundle.
func (t TLSConfig) MutualTLS() bool { return t.Enabled() && t.ClientCAFile != "" }

// HTTPConfig configures the metrics/probes sidecar HTTP server. gRPC has no
// metrics endpoint, so Prometheus /metrics and the k8s probes live here.
type HTTPConfig struct {
	// Addr is the sidecar listen address, e.g. ":8080".
	Addr string
	// ReadHeaderTimeout bounds how long reading request headers may take.
	ReadHeaderTimeout time.Duration
}

// AuthConfig gates the bearer-token interceptor. With Enabled false the service
// runs in a local/dev mode that injects a synthetic principal so it boots
// offline without an identity provider. With Enabled true a static shared token
// is required (the reference uses a static token to stay dependency-free; a real
// build swaps in a JWKS verifier behind the same interceptor seam).
type AuthConfig struct {
	// Enabled gates the auth interceptor. When false the service runs without
	// authentication (local/dev); when true a valid bearer token is required.
	Enabled bool
	// Token is the expected bearer token when Enabled is true.
	Token string
}

// TelemetryConfig configures structured logging, tracing, and the metrics seam.
type TelemetryConfig struct {
	// LogLevel is the minimum slog level.
	LogLevel slog.Level
	// LogJSON selects JSON output (production) over text (local dev).
	LogJSON bool
	// OTLPEndpoint is the OTLP/HTTP trace collector endpoint (host:port, no
	// scheme). Empty disables span export: the service installs a never-sampling
	// provider so it runs offline.
	OTLPEndpoint string
	// OTLPInsecure sends spans over plaintext HTTP rather than TLS.
	OTLPInsecure bool
	// TraceSampleRatio is the head-based parent sampling ratio in [0,1].
	TraceSampleRatio float64
}

// Default values. Kept as named constants so the defaults are a single,
// reviewable source of truth rather than scattered literals.
const (
	defaultGRPCAddr          = ":9090"
	defaultMaxRecvMsgBytes   = 4 << 20 // 4 MiB (grpc-go default)
	defaultHandlerTimeout    = 15 * time.Second
	defaultConnTimeout       = 10 * time.Second
	defaultHTTPAddr          = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultShutdownGrace     = 15 * time.Second
	defaultTraceSampleRatio  = 1.0
)

// Load reads configuration from flags and the environment, applies defaults,
// and validates the result. Precedence is flags > environment > defaults: each
// flag's default is seeded from the environment (or the hard-coded default), so
// an explicit flag wins, an unset flag falls back to the env value, and an
// unset env falls back to the default.
//
// args are the process arguments excluding the program name (os.Args[1:]).
func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("examplegrpc", flag.ContinueOnError)

	// Seed each flag default from the environment (or the hard-coded default). A
	// malformed env value is fail-fast: env collects the parse error keyed by the
	// offending variable and Load aborts before any flag parsing, naming the bad
	// key — it is never silently coerced to the default.
	env := newEnvReader()

	grpcAddr := fs.String("grpc-addr", env.string("GRPC_ADDR", defaultGRPCAddr), "gRPC listen address")
	maxRecvMsgBytes := fs.Int("grpc-max-recv-msg-bytes", env.int("GRPC_MAX_RECV_MSG_BYTES", defaultMaxRecvMsgBytes), "max inbound gRPC message size in bytes")
	handlerTimeout := fs.Duration("grpc-handler-timeout", env.duration("GRPC_HANDLER_TIMEOUT", defaultHandlerTimeout), "deadline-guard ceiling for RPCs with no client deadline")
	connTimeout := fs.Duration("grpc-conn-timeout", env.duration("GRPC_CONN_TIMEOUT", defaultConnTimeout), "max time for a new connection to become ready")

	tlsCertFile := fs.String("grpc-tls-cert-file", env.string("GRPC_TLS_CERT_FILE", ""), "PEM server certificate path (enables TLS with key; empty = insecure local/dev)")
	tlsKeyFile := fs.String("grpc-tls-key-file", env.string("GRPC_TLS_KEY_FILE", ""), "PEM server private key path (required with cert)")
	tlsClientCAFile := fs.String("grpc-tls-client-ca-file", env.string("GRPC_TLS_CLIENT_CA_FILE", ""), "PEM client-CA bundle path (enables mTLS client-cert verification)")

	httpAddr := fs.String("http-addr", env.string("HTTP_ADDR", defaultHTTPAddr), "metrics/probes sidecar listen address")
	readHeaderTimeout := fs.Duration("http-read-header-timeout", env.duration("HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout), "sidecar HTTP read-header timeout")

	logLevel := fs.String("log-level", env.string("LOG_LEVEL", "info"), "log level (debug|info|warn|error)")
	logJSON := fs.Bool("log-json", env.bool("LOG_JSON", true), "emit JSON logs (false for text)")
	otlpEndpoint := fs.String("otlp-endpoint", env.string("OTLP_ENDPOINT", ""), "OTLP/HTTP trace endpoint host:port (empty disables span export)")
	otlpInsecure := fs.Bool("otlp-insecure", env.bool("OTLP_INSECURE", false), "send spans over plaintext HTTP instead of TLS")
	traceSampleRatio := fs.Float64("trace-sample-ratio", env.float64("TRACE_SAMPLE_RATIO", defaultTraceSampleRatio), "head-based trace sampling ratio in [0,1]")

	authEnabled := fs.Bool("auth-enabled", env.bool("AUTH_ENABLED", false), "require a valid bearer token on API calls (off for local/dev)")
	authToken := fs.String("auth-token", env.string("AUTH_TOKEN", ""), "expected bearer token when auth is enabled")

	shutdownGrace := fs.Duration("shutdown-grace", env.duration("SHUTDOWN_GRACE", defaultShutdownGrace), "graceful shutdown budget")

	// Abort before parsing flags if any env value was malformed.
	if err := env.err(); err != nil {
		return Config{}, err
	}

	if err := fs.Parse(args); err != nil {
		return Config{}, fmt.Errorf("parse flags: %w", err)
	}

	level, err := parseLevel(*logLevel)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		GRPC: GRPCConfig{
			Addr:            *grpcAddr,
			MaxRecvMsgBytes: *maxRecvMsgBytes,
			HandlerTimeout:  *handlerTimeout,
			ConnTimeout:     *connTimeout,
			TLS: TLSConfig{
				CertFile:     *tlsCertFile,
				KeyFile:      *tlsKeyFile,
				ClientCAFile: *tlsClientCAFile,
			},
		},
		HTTP: HTTPConfig{
			Addr:              *httpAddr,
			ReadHeaderTimeout: *readHeaderTimeout,
		},
		Telemetry: TelemetryConfig{
			LogLevel:         level,
			LogJSON:          *logJSON,
			OTLPEndpoint:     *otlpEndpoint,
			OTLPInsecure:     *otlpInsecure,
			TraceSampleRatio: *traceSampleRatio,
		},
		Auth: AuthConfig{
			Enabled: *authEnabled,
			Token:   *authToken,
		},
		ShutdownGrace: *shutdownGrace,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate enforces the invariants that must hold before the process opens
// listeners or external clients. It reports the first violation with an
// actionable message; it never silently corrects a bad value.
func (c Config) Validate() error {
	if c.GRPC.Addr == "" {
		return errors.New("config: GRPC_ADDR must not be empty")
	}
	if c.GRPC.MaxRecvMsgBytes <= 0 {
		return fmt.Errorf("config: GRPC_MAX_RECV_MSG_BYTES must be positive, got %d", c.GRPC.MaxRecvMsgBytes)
	}
	if c.GRPC.HandlerTimeout <= 0 {
		return fmt.Errorf("config: GRPC_HANDLER_TIMEOUT must be positive, got %s", c.GRPC.HandlerTimeout)
	}
	if c.GRPC.ConnTimeout <= 0 {
		return fmt.Errorf("config: GRPC_CONN_TIMEOUT must be positive, got %s", c.GRPC.ConnTimeout)
	}
	// TLS cert and key are required together: a half-configured pair would either
	// fail at handshake or silently fall back to insecure, so it is fail-fast.
	if (c.GRPC.TLS.CertFile == "") != (c.GRPC.TLS.KeyFile == "") {
		return errors.New("config: GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE must be set together")
	}
	// A client-CA bundle only makes sense with server TLS; requiring client certs
	// without serving TLS is a misconfiguration.
	if c.GRPC.TLS.ClientCAFile != "" && !c.GRPC.TLS.Enabled() {
		return errors.New("config: GRPC_TLS_CLIENT_CA_FILE (mTLS) requires GRPC_TLS_CERT_FILE and GRPC_TLS_KEY_FILE")
	}
	if c.HTTP.Addr == "" {
		return errors.New("config: HTTP_ADDR must not be empty")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("config: HTTP_READ_HEADER_TIMEOUT must be positive, got %s", c.HTTP.ReadHeaderTimeout)
	}
	if c.ShutdownGrace <= 0 {
		return fmt.Errorf("config: SHUTDOWN_GRACE must be positive, got %s", c.ShutdownGrace)
	}
	if c.Telemetry.TraceSampleRatio < 0 || c.Telemetry.TraceSampleRatio > 1 {
		return fmt.Errorf("config: TRACE_SAMPLE_RATIO must be in [0,1], got %v", c.Telemetry.TraceSampleRatio)
	}
	// When auth is enabled the token is required: booting with an empty token
	// would accept or reject calls unpredictably, so it is a fail-fast error.
	if c.Auth.Enabled && c.Auth.Token == "" {
		return errors.New("config: AUTH_TOKEN must be set when AUTH_ENABLED=true")
	}
	return nil
}

func parseLevel(s string) (slog.Level, error) {
	var level slog.Level
	// slog.Level implements encoding.TextUnmarshaler and accepts
	// debug/info/warn/error (case-insensitive).
	if err := level.UnmarshalText([]byte(s)); err != nil {
		return 0, fmt.Errorf("config: invalid LOG_LEVEL %q: %w", s, err)
	}
	return level, nil
}

// envReader reads typed values from the environment and accumulates parse
// errors instead of swallowing them. A malformed value is a fail-fast condition
// per golang/foundations/configuration.md: each getter records an actionable,
// key-named error and returns the fallback so default seeding can proceed, then
// Load checks err() and aborts before opening listeners.
type envReader struct {
	errs []error
}

func newEnvReader() *envReader { return &envReader{} }

// err returns the joined parse errors, or nil if every read succeeded.
func (e *envReader) err() error {
	if len(e.errs) == 0 {
		return nil
	}
	return fmt.Errorf("config: invalid environment: %w", errors.Join(e.errs...))
}

func (e *envReader) string(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func (e *envReader) int(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s must be an integer, got %q", key, v))
		return fallback
	}
	return n
}

func (e *envReader) duration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s must be a duration like 15s or 1m, got %q", key, v))
		return fallback
	}
	return d
}

func (e *envReader) float64(key string, fallback float64) float64 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s must be a number, got %q", key, v))
		return fallback
	}
	return f
}

func (e *envReader) bool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s must be a boolean (true/false/1/0), got %q", key, v))
		return fallback
	}
	return b
}
