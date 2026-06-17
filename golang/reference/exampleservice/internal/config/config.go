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
	// HTTP holds the listener and server-hardening settings.
	HTTP HTTPConfig
	// Database holds the connection string and pool sizing.
	Database DatabaseConfig
	// Telemetry holds logging and metrics configuration.
	Telemetry TelemetryConfig
	// ShutdownGrace bounds ordered shutdown. It must exceed worst-case
	// in-flight work and stay under the platform termination grace.
	ShutdownGrace time.Duration
}

// HTTPConfig configures the HTTP server and its hardening timeouts.
type HTTPConfig struct {
	// Addr is the listen address, e.g. ":8080".
	Addr string
	// ReadHeaderTimeout bounds how long reading request headers may take.
	ReadHeaderTimeout time.Duration
	// ReadTimeout bounds reading the entire request including the body.
	ReadTimeout time.Duration
	// WriteTimeout bounds writing the response. Disable only for streaming.
	WriteTimeout time.Duration
	// IdleTimeout bounds how long an idle keep-alive connection lives.
	IdleTimeout time.Duration
	// MaxBodyBytes caps non-streaming request bodies before decoding.
	MaxBodyBytes int64
}

// DatabaseConfig configures the database pool. All four pool limits are set
// explicitly from config per golang/services/database.md; the defaults here
// are deliberate, not the database/sql zero values.
type DatabaseConfig struct {
	// DSN is the data source name. Empty selects the in-memory store.
	DSN string
	// MaxOpenConns caps total open connections (server capacity / instances).
	MaxOpenConns int
	// MaxIdleConns is the idle floor; must be <= MaxOpenConns.
	MaxIdleConns int
	// ConnMaxLifetime bounds connection age (required behind LB/proxy/failover).
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime reaps idle connections so the pool shrinks under low load.
	ConnMaxIdleTime time.Duration
}

// TelemetryConfig configures structured logging and the metrics seam.
type TelemetryConfig struct {
	// LogLevel is the minimum slog level.
	LogLevel slog.Level
	// LogJSON selects JSON output (production) over text (local dev).
	LogJSON bool
}

// Default values. Kept as named constants so the defaults are a single,
// reviewable source of truth rather than scattered literals.
const (
	defaultAddr              = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultMaxBodyBytes      = 1 << 20 // 1 MiB
	defaultMaxOpenConns      = 25
	defaultMaxIdleConns      = 25
	defaultConnMaxLifetime   = 30 * time.Minute
	defaultConnMaxIdleTime   = 5 * time.Minute
	defaultShutdownGrace     = 15 * time.Second
)

// Load reads configuration from flags and the environment, applies defaults,
// and validates the result. Precedence is flags > environment > defaults: each
// flag's default is seeded from the environment (or the hard-coded default), so
// an explicit flag wins, an unset flag falls back to the env value, and an
// unset env falls back to the default.
//
// args are the process arguments excluding the program name (os.Args[1:]).
// Passing them in keeps Load testable without mutating global flag state.
func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("exampleservice", flag.ContinueOnError)

	// Seed each flag default from the environment (or the hard-coded default).
	// A malformed env value is fail-fast: env collects the parse error keyed by
	// the offending variable and Load aborts before any flag parsing, naming the
	// bad key — it is never silently coerced to the default. This mirrors the
	// LOG_LEVEL path, which has always rejected garbage rather than defaulting.
	env := newEnvReader()

	addr := fs.String("http-addr", env.string("HTTP_ADDR", defaultAddr), "HTTP listen address")
	readHeaderTimeout := fs.Duration("http-read-header-timeout", env.duration("HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout), "HTTP read-header timeout")
	readTimeout := fs.Duration("http-read-timeout", env.duration("HTTP_READ_TIMEOUT", defaultReadTimeout), "HTTP read timeout")
	writeTimeout := fs.Duration("http-write-timeout", env.duration("HTTP_WRITE_TIMEOUT", defaultWriteTimeout), "HTTP write timeout")
	idleTimeout := fs.Duration("http-idle-timeout", env.duration("HTTP_IDLE_TIMEOUT", defaultIdleTimeout), "HTTP idle timeout")
	maxBodyBytes := fs.Int64("http-max-body-bytes", env.int64("HTTP_MAX_BODY_BYTES", defaultMaxBodyBytes), "max request body size in bytes")

	dsn := fs.String("db-dsn", env.string("DB_DSN", ""), "database DSN (empty uses the in-memory store)")
	maxOpenConns := fs.Int("db-max-open-conns", env.int("DB_MAX_OPEN_CONNS", defaultMaxOpenConns), "max open DB connections")
	maxIdleConns := fs.Int("db-max-idle-conns", env.int("DB_MAX_IDLE_CONNS", defaultMaxIdleConns), "max idle DB connections")
	connMaxLifetime := fs.Duration("db-conn-max-lifetime", env.duration("DB_CONN_MAX_LIFETIME", defaultConnMaxLifetime), "max DB connection lifetime")
	connMaxIdleTime := fs.Duration("db-conn-max-idle-time", env.duration("DB_CONN_MAX_IDLE_TIME", defaultConnMaxIdleTime), "max DB connection idle time")

	logLevel := fs.String("log-level", env.string("LOG_LEVEL", "info"), "log level (debug|info|warn|error)")
	logJSON := fs.Bool("log-json", env.bool("LOG_JSON", true), "emit JSON logs (false for text)")

	shutdownGrace := fs.Duration("shutdown-grace", env.duration("SHUTDOWN_GRACE", defaultShutdownGrace), "graceful shutdown budget")

	// Abort before parsing flags if any env value was malformed: a bad default
	// would otherwise be silently masked by an explicit flag or, worse, accepted.
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
		HTTP: HTTPConfig{
			Addr:              *addr,
			ReadHeaderTimeout: *readHeaderTimeout,
			ReadTimeout:       *readTimeout,
			WriteTimeout:      *writeTimeout,
			IdleTimeout:       *idleTimeout,
			MaxBodyBytes:      *maxBodyBytes,
		},
		Database: DatabaseConfig{
			DSN:             *dsn,
			MaxOpenConns:    *maxOpenConns,
			MaxIdleConns:    *maxIdleConns,
			ConnMaxLifetime: *connMaxLifetime,
			ConnMaxIdleTime: *connMaxIdleTime,
		},
		Telemetry: TelemetryConfig{
			LogLevel: level,
			LogJSON:  *logJSON,
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
	if c.HTTP.Addr == "" {
		return errors.New("config: HTTP_ADDR must not be empty")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("config: HTTP_READ_HEADER_TIMEOUT must be positive, got %s", c.HTTP.ReadHeaderTimeout)
	}
	if c.HTTP.MaxBodyBytes <= 0 {
		return fmt.Errorf("config: HTTP_MAX_BODY_BYTES must be positive, got %d", c.HTTP.MaxBodyBytes)
	}
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("config: DB_MAX_OPEN_CONNS must be positive, got %d", c.Database.MaxOpenConns)
	}
	// The pool invariant from golang/services/database.md: an idle floor above
	// the open cap is nonsensical and database/sql would silently clamp it.
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("config: DB_MAX_IDLE_CONNS (%d) must be <= DB_MAX_OPEN_CONNS (%d)", c.Database.MaxIdleConns, c.Database.MaxOpenConns)
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("config: DB_CONN_MAX_LIFETIME must be positive, got %s", c.Database.ConnMaxLifetime)
	}
	if c.ShutdownGrace <= 0 {
		return fmt.Errorf("config: SHUTDOWN_GRACE must be positive, got %s", c.ShutdownGrace)
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
// errors instead of swallowing them. A malformed value (e.g. DB_MAX_OPEN_CONNS=abc)
// is a fail-fast condition per golang/foundations/configuration.md ("no silent
// fallback when a value is malformed"): each getter records an actionable,
// key-named error and returns the fallback so default seeding can proceed, then
// Load checks err() and aborts before opening listeners. The fallback is used
// only to keep building the flag set; it is never the accepted config when an
// error was recorded.
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

func (e *envReader) int64(key string, fallback int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		e.errs = append(e.errs, fmt.Errorf("%s must be a 64-bit integer, got %q", key, v))
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
