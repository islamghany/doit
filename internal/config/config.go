package config

import (
	"fmt"

	"github.com/islamghany/enfl"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
)

type DatabaseConfig struct {
	Name            string `env:"NAME" flag:"db_name" required:"true"`
	User            string `env:"USER" flag:"db_user" required:"true"`
	Password        string `env:"PASSWORD" flag:"db_password"`
	DisableTLS      bool   `env:"DISABLE_TLS" flag:"db_disable_tls"`
	Host            string `env:"HOST" flag:"db_host" required:"true"`
	Port            int    `env:"PORT" flag:"db_port" required:"true"`
	MaxOpenConns    int    `env:"MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int    `env:"MAX_IDLE_CONNS" default:"25"`
	ConnMaxLifetime int    `env:"CONN_MAX_LIFETIME" default:"300"` // 5 minutes in seconds
}

type RedisConfig struct {
	Addr     string `env:"ADDR" flag:"redis_addr" required:"true"`
	Password string `env:"PASSWORD" flag:"redis_password"`
	Database int    `env:"DB" flag:"redis_db" default:"0"`
}

type JWTConfig struct {
	Secret          string `env:"SECRET" flag:"jwt_secret" required:"true"`
	RefreshSecret   string `env:"REFRESH_SECRET" flag:"refresh_secret" required:"true"`
	AccessTokenExp  int    `env:"ACCESS_TOKEN_EXP" flag:"access_token_exp" required:"true"`
	RefreshTokenExp int    `env:"REFRESH_TOKEN_EXP" flag:"refresh_token_exp" required:"true"`
}

type ServerConfig struct {
	Host            string `env:"HOST" flag:"host" required:"true"`
	Port            int    `env:"PORT" flag:"port" required:"true"`
	ReadTimeout     int    `env:"READ_TIMEOUT" flag:"read_timeout" default:"10"`         // seconds
	WriteTimeout    int    `env:"WRITE_TIMEOUT" flag:"write_timeout" default:"10"`       // seconds
	IdleTimeout     int    `env:"IDLE_TIMEOUT" flag:"idle_timeout" default:"120"`        // seconds
	ShutdownTimeout int    `env:"SHUTDOWN_TIMEOUT" flag:"shutdown_timeout" default:"25"` // seconds
}

type AppConfig struct {
	Environment Environment `env:"ENVIRONMENT" flag:"environment" default:"development"`
	LogLevel    string      `env:"LOG_LEVEL" flag:"log_level" default:"info"`
}

type SecurityConfig struct {
	EnableCORS        bool     `env:"ENABLE_CORS" flag:"enable_cors" default:"true"`
	AllowedOrigins    []string `env:"ALLOWED_ORIGINS" flag:"allowed_origins" default:"*"`
	RateLimitEnabled  bool     `env:"RATE_LIMIT_ENABLED" flag:"rate_limit_enabled" default:"true"`
	MaxRequestsPerMin int      `env:"MAX_REQUESTS_PER_MIN" flag:"max_requests_per_min" default:"100"`
	EnableTLS         bool     `env:"ENABLE_TLS" flag:"enable_tls" default:"false"`
	TLSCertFile       string   `env:"TLS_CERT_FILE" flag:"tls_cert_file"`
	TLSKeyFile        string   `env:"TLS_KEY_FILE" flag:"tls_key_file"`
}

type Config struct {
	Server   ServerConfig   `prefix:"SERVER_"`
	App      AppConfig      `prefix:"APP_"`
	Database DatabaseConfig `prefix:"DB_"`
	JWT      JWTConfig      `prefix:"JWT_"`
	Security SecurityConfig `prefix:"SECURITY_"`
	Redis    RedisConfig    `prefix:"REDIS_"`
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Server validation
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d (must be between 1-65535)", c.Server.Port)
	}

	// App validation
	validEnvs := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
	}
	if !validEnvs[string(c.App.Environment)] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, or production)", c.App.Environment)
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.App.LogLevel] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.App.LogLevel)
	}

	// Add more validation when you uncomment other configs

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) DevPrint() {
	if c.IsDevelopment() {
		// Print the config in a pretty format
		fmt.Printf("%+v", c)
	}
}

func LoadConfig() (*Config, error) {
	var cfg Config
	if err := enfl.Load(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}
