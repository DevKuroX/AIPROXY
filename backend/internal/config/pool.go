package config

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig holds connection pool configuration for PostgreSQL and HTTP clients.
type PoolConfig struct {
	// PostgreSQL pool settings
	PGMaxConns        int32         `json:"pg_max_conns"`
	PGMinConns        int32         `json:"pg_min_conns"`
	PGMaxConnLifetime time.Duration `json:"pg_max_conn_lifetime"`
	PGMaxConnIdleTime time.Duration `json:"pg_max_conn_idle_time"`
	PGHealthCheckPeriod time.Duration `json:"pg_health_check_period"`

	// HTTP client pool settings
	HTTPMaxIdleConns        int           `json:"http_max_idle_conns"`
	HTTPMaxIdleConnsPerHost int           `json:"http_max_idle_conns_per_host"`
	HTTPIdleConnTimeout     time.Duration `json:"http_idle_conn_timeout"`
	HTTPTimeout             time.Duration `json:"http_timeout"`
	HTTPKeepAlive           time.Duration `json:"http_keep_alive"`
	HTTPMaxConnsPerHost     int           `json:"http_max_conns_per_host"`
}

// DefaultPoolConfig returns production-ready defaults optimized for high concurrency.
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		// PostgreSQL: optimized for connection reuse under load
		PGMaxConns:          50,
		PGMinConns:          5,
		PGMaxConnLifetime:   30 * time.Minute,
		PGMaxConnIdleTime:   5 * time.Minute,
		PGHealthCheckPeriod: 1 * time.Minute,

		// HTTP: optimized for SSE streaming with many concurrent connections
		HTTPMaxIdleConns:        500,
		HTTPMaxIdleConnsPerHost: 100,
		HTTPIdleConnTimeout:     90 * time.Second,
		HTTPTimeout:             30 * time.Second,
		HTTPKeepAlive:           30 * time.Second,
		HTTPMaxConnsPerHost:     0, // 0 = unlimited, controlled by MaxIdleConnsPerHost
	}
}

// LoadPoolConfigFromEnv loads pool configuration from environment variables.
func LoadPoolConfigFromEnv() *PoolConfig {
	cfg := DefaultPoolConfig()

	if v := getEnvInt("PG_MAX_CONNS", 0); v > 0 {
		cfg.PGMaxConns = int32(v)
	}
	if v := getEnvInt("PG_MIN_CONNS", 0); v > 0 {
		cfg.PGMinConns = int32(v)
	}
	if v := getEnvDuration("PG_MAX_CONN_LIFETIME", 0); v > 0 {
		cfg.PGMaxConnLifetime = v
	}
	if v := getEnvDuration("PG_MAX_CONN_IDLE_TIME", 0); v > 0 {
		cfg.PGMaxConnIdleTime = v
	}
	if v := getEnvDuration("PG_HEALTH_CHECK_PERIOD", 0); v > 0 {
		cfg.PGHealthCheckPeriod = v
	}

	if v := getEnvInt("HTTP_MAX_IDLE_CONNS", 0); v > 0 {
		cfg.HTTPMaxIdleConns = v
	}
	if v := getEnvInt("HTTP_MAX_IDLE_CONNS_PER_HOST", 0); v > 0 {
		cfg.HTTPMaxIdleConnsPerHost = v
	}
	if v := getEnvDuration("HTTP_IDLE_CONN_TIMEOUT", 0); v > 0 {
		cfg.HTTPIdleConnTimeout = v
	}
	if v := getEnvDuration("HTTP_TIMEOUT", 0); v > 0 {
		cfg.HTTPTimeout = v
	}
	if v := getEnvDuration("HTTP_KEEP_ALIVE", 0); v > 0 {
		cfg.HTTPKeepAlive = v
	}
	if v := getEnvInt("HTTP_MAX_CONNS_PER_HOST", -1); v >= 0 {
		cfg.HTTPMaxConnsPerHost = v
	}

	return cfg
}

// NewPgxPool creates a configured PostgreSQL connection pool.
func NewPgxPool(ctx context.Context, dbURL string, cfg *PoolConfig) (*pgxpool.Pool, error) {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	poolConfig.MaxConns = cfg.PGMaxConns
	poolConfig.MinConns = cfg.PGMinConns
	poolConfig.MaxConnLifetime = cfg.PGMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.PGMaxConnIdleTime
	poolConfig.HealthCheckPeriod = cfg.PGHealthCheckPeriod

	// Set reasonable connection establishment timeout
	poolConfig.ConnConfig.ConnectTimeout = 10 * time.Second

	return pgxpool.NewWithConfig(ctx, poolConfig)
}

// NewHTTPClient creates a configured HTTP client optimized for upstream API calls.
// This client is suitable for both regular requests and SSE streaming.
func NewHTTPClient(cfg *PoolConfig) *http.Client {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: cfg.HTTPKeepAlive,
		}).DialContext,
		MaxIdleConns:          cfg.HTTPMaxIdleConns,
		MaxIdleConnsPerHost:   cfg.HTTPMaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.HTTPIdleConnTimeout,
		MaxConnsPerHost:       cfg.HTTPMaxConnsPerHost,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		// Enable HTTP/2 for better performance with multiplexing
		ForceAttemptHTTP2: true,
	}

	return &http.Client{
		Transport: transport,
		Timeout:   cfg.HTTPTimeout,
	}
}

// NewStreamingHTTPClient creates an HTTP client optimized for SSE streaming.
// The timeout is set higher to allow long-lived streaming connections.
func NewStreamingHTTPClient(cfg *PoolConfig) *http.Client {
	if cfg == nil {
		cfg = DefaultPoolConfig()
	}

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: cfg.HTTPKeepAlive,
		}).DialContext,
		MaxIdleConns:          cfg.HTTPMaxIdleConns,
		MaxIdleConnsPerHost:   cfg.HTTPMaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.HTTPIdleConnTimeout,
		MaxConnsPerHost:       cfg.HTTPMaxConnsPerHost,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	}

	// No timeout for streaming - let the transport handle idle connections
	return &http.Client{
		Transport: transport,
	}
}

// getEnvDuration parses a duration from an environment variable.
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := getEnv(key, ""); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
