// Package config loads, defaults, and validates all process configuration in
// one place. Precedence is flags > environment > hard-coded defaults, per
// golang/foundations/configuration.md. Validation is fail-fast: Load returns a
// fully validated Config or an actionable error, and main aborts before
// connecting to the broker or opening the probe listener.
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
	// HTTP holds the probe/metrics sidecar listener settings.
	HTTP HTTPConfig
	// Consumer holds the consume-loop retry/backoff/DLQ policy.
	Consumer ConsumerConfig
	// Outbox holds the transactional-outbox relay settings.
	Outbox OutboxConfig
	// Telemetry holds logging, tracing, and metrics configuration.
	Telemetry TelemetryConfig
	// Topic is the broker topic/subject the consumer subscribes to and the
	// outbox relay publishes to. Broker selection is an ADR decision; the
	// reference uses an in-memory broker so this is just a logical name.
	Topic string
	// ShutdownGrace bounds the ordered graceful drain on shutdown. It must
	// exceed worst-case in-flight message processing and stay under the platform
	// termination grace.
	ShutdownGrace time.Duration
}

// HTTPConfig configures the probe/metrics sidecar HTTP server and its hardening
// timeouts. The worker is not an HTTP service; this listener serves only
// /livez, /readyz, and /metrics so the platform can probe it and Prometheus can
// scrape it.
type HTTPConfig struct {
	// Addr is the listen address, e.g. ":8080".
	Addr string
	// ReadHeaderTimeout bounds how long reading request headers may take.
	ReadHeaderTimeout time.Duration
	// ReadTimeout bounds reading the entire request including the body.
	ReadTimeout time.Duration
	// WriteTimeout bounds writing the response.
	WriteTimeout time.Duration
	// IdleTimeout bounds how long an idle keep-alive connection lives.
	IdleTimeout time.Duration
}

// ConsumerConfig configures the consume loop's bounded-retry policy. Retries use
// exponential backoff with full jitter computed from the injected clock, per
// golang/services/eventing-and-messaging.md (### Retries And Dead-Letter
// Behavior). After MaxAttempts the message is dead-lettered.
type ConsumerConfig struct {
	// MaxAttempts is the total number of delivery attempts (including the first)
	// before a message is dead-lettered. Must be >= 1.
	MaxAttempts int
	// BaseBackoff is the first retry's backoff ceiling; subsequent attempts
	// double it (capped at MaxBackoff) before full jitter is applied.
	BaseBackoff time.Duration
	// MaxBackoff caps the exponential backoff ceiling.
	MaxBackoff time.Duration
}

// OutboxConfig configures the transactional-outbox relay loop.
type OutboxConfig struct {
	// PollInterval is how often the relay scans the outbox store for pending
	// rows to publish. The reference relay also runs on demand in tests.
	PollInterval time.Duration
	// BatchSize caps how many pending rows the relay claims per scan.
	BatchSize int
}

// TelemetryConfig configures structured logging, tracing, and the metrics seam.
type TelemetryConfig struct {
	// LogLevel is the minimum slog level.
	LogLevel slog.Level
	// LogJSON selects JSON output (production) over text (local dev).
	LogJSON bool
	// OTLPEndpoint is the OTLP/HTTP trace collector endpoint (host:port, no
	// scheme), e.g. "otel-collector:4318". Empty disables span export: the
	// worker installs a never-sampling provider so it runs offline.
	OTLPEndpoint string
	// OTLPInsecure sends spans over plaintext HTTP rather than TLS. Only for
	// in-cluster collectors / local development.
	OTLPInsecure bool
	// TraceSampleRatio is the head-based parent sampling ratio in [0,1].
	TraceSampleRatio float64
}

// Default values. Kept as named constants so the defaults are a single,
// reviewable source of truth rather than scattered literals.
const (
	defaultAddr              = ":8080"
	defaultReadHeaderTimeout = 5 * time.Second
	defaultReadTimeout       = 15 * time.Second
	defaultWriteTimeout      = 15 * time.Second
	defaultIdleTimeout       = 60 * time.Second
	defaultMaxAttempts       = 5
	defaultBaseBackoff       = 100 * time.Millisecond
	defaultMaxBackoff        = 30 * time.Second
	defaultOutboxPoll        = time.Second
	defaultOutboxBatch       = 100
	defaultTopic             = "widget.events"
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
// Passing them in keeps Load testable without mutating global flag state.
func Load(args []string) (Config, error) {
	fs := flag.NewFlagSet("exampleworker", flag.ContinueOnError)

	// Seed each flag default from the environment (or the hard-coded default). A
	// malformed env value is fail-fast: env collects the parse error keyed by the
	// offending variable and Load aborts before any flag parsing, naming the bad
	// key — it is never silently coerced to the default.
	env := newEnvReader()

	addr := fs.String("http-addr", env.string("HTTP_ADDR", defaultAddr), "probe/metrics sidecar listen address")
	readHeaderTimeout := fs.Duration("http-read-header-timeout", env.duration("HTTP_READ_HEADER_TIMEOUT", defaultReadHeaderTimeout), "HTTP read-header timeout")
	readTimeout := fs.Duration("http-read-timeout", env.duration("HTTP_READ_TIMEOUT", defaultReadTimeout), "HTTP read timeout")
	writeTimeout := fs.Duration("http-write-timeout", env.duration("HTTP_WRITE_TIMEOUT", defaultWriteTimeout), "HTTP write timeout")
	idleTimeout := fs.Duration("http-idle-timeout", env.duration("HTTP_IDLE_TIMEOUT", defaultIdleTimeout), "HTTP idle timeout")

	maxAttempts := fs.Int("consumer-max-attempts", env.int("CONSUMER_MAX_ATTEMPTS", defaultMaxAttempts), "total delivery attempts before dead-lettering")
	baseBackoff := fs.Duration("consumer-base-backoff", env.duration("CONSUMER_BASE_BACKOFF", defaultBaseBackoff), "base retry backoff (doubles per attempt)")
	maxBackoff := fs.Duration("consumer-max-backoff", env.duration("CONSUMER_MAX_BACKOFF", defaultMaxBackoff), "max retry backoff ceiling")

	outboxPoll := fs.Duration("outbox-poll-interval", env.duration("OUTBOX_POLL_INTERVAL", defaultOutboxPoll), "how often the outbox relay scans for pending rows")
	outboxBatch := fs.Int("outbox-batch-size", env.int("OUTBOX_BATCH_SIZE", defaultOutboxBatch), "max pending rows the outbox relay claims per scan")

	topic := fs.String("topic", env.string("TOPIC", defaultTopic), "broker topic/subject for the consumer and outbox relay")

	logLevel := fs.String("log-level", env.string("LOG_LEVEL", "info"), "log level (debug|info|warn|error)")
	logJSON := fs.Bool("log-json", env.bool("LOG_JSON", true), "emit JSON logs (false for text)")
	otlpEndpoint := fs.String("otlp-endpoint", env.string("OTLP_ENDPOINT", ""), "OTLP/HTTP trace endpoint host:port (empty disables span export)")
	otlpInsecure := fs.Bool("otlp-insecure", env.bool("OTLP_INSECURE", false), "send spans over plaintext HTTP instead of TLS")
	traceSampleRatio := fs.Float64("trace-sample-ratio", env.float64("TRACE_SAMPLE_RATIO", defaultTraceSampleRatio), "head-based trace sampling ratio in [0,1]")

	shutdownGrace := fs.Duration("shutdown-grace", env.duration("SHUTDOWN_GRACE", defaultShutdownGrace), "graceful drain budget")

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
		},
		Consumer: ConsumerConfig{
			MaxAttempts: *maxAttempts,
			BaseBackoff: *baseBackoff,
			MaxBackoff:  *maxBackoff,
		},
		Outbox: OutboxConfig{
			PollInterval: *outboxPoll,
			BatchSize:    *outboxBatch,
		},
		Telemetry: TelemetryConfig{
			LogLevel:         level,
			LogJSON:          *logJSON,
			OTLPEndpoint:     *otlpEndpoint,
			OTLPInsecure:     *otlpInsecure,
			TraceSampleRatio: *traceSampleRatio,
		},
		Topic:         *topic,
		ShutdownGrace: *shutdownGrace,
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate enforces the invariants that must hold before the process connects
// to the broker or opens the probe listener. It reports the first violation
// with an actionable message; it never silently corrects a bad value.
func (c Config) Validate() error {
	if c.HTTP.Addr == "" {
		return errors.New("config: HTTP_ADDR must not be empty")
	}
	if c.HTTP.ReadHeaderTimeout <= 0 {
		return fmt.Errorf("config: HTTP_READ_HEADER_TIMEOUT must be positive, got %s", c.HTTP.ReadHeaderTimeout)
	}
	if c.Topic == "" {
		return errors.New("config: TOPIC must not be empty")
	}
	if c.Consumer.MaxAttempts < 1 {
		return fmt.Errorf("config: CONSUMER_MAX_ATTEMPTS must be >= 1, got %d", c.Consumer.MaxAttempts)
	}
	if c.Consumer.BaseBackoff <= 0 {
		return fmt.Errorf("config: CONSUMER_BASE_BACKOFF must be positive, got %s", c.Consumer.BaseBackoff)
	}
	if c.Consumer.MaxBackoff < c.Consumer.BaseBackoff {
		return fmt.Errorf("config: CONSUMER_MAX_BACKOFF (%s) must be >= CONSUMER_BASE_BACKOFF (%s)", c.Consumer.MaxBackoff, c.Consumer.BaseBackoff)
	}
	if c.Outbox.PollInterval <= 0 {
		return fmt.Errorf("config: OUTBOX_POLL_INTERVAL must be positive, got %s", c.Outbox.PollInterval)
	}
	if c.Outbox.BatchSize < 1 {
		return fmt.Errorf("config: OUTBOX_BATCH_SIZE must be >= 1, got %d", c.Outbox.BatchSize)
	}
	if c.ShutdownGrace <= 0 {
		return fmt.Errorf("config: SHUTDOWN_GRACE must be positive, got %s", c.ShutdownGrace)
	}
	if c.Telemetry.TraceSampleRatio < 0 || c.Telemetry.TraceSampleRatio > 1 {
		return fmt.Errorf("config: TRACE_SAMPLE_RATIO must be in [0,1], got %v", c.Telemetry.TraceSampleRatio)
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
// Load checks err() and aborts before connecting to the broker. The fallback is
// used only to keep building the flag set; it is never the accepted config when
// an error was recorded.
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
