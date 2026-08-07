package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	ConfigType        = "yml"
	DefaultConfigPath = "/run/secrets"
	DefaultConfigName = "cluer"
	EnvConfigPath     = "CONFIG_PATH"
	EnvConfigName     = "CONFIG_NAME"
	EnvPrefix         = "cluer"
)

type (
	// Main struct for top-level config
	Config struct {
		ServerConfig   *ServerConfig   `mapstructure:"server" validate:"required"`
		PostgresConfig *PostgresConfig `mapstructure:"postgres" validate:"omitempty"`
		AuthConfig     *AuthConfig     `mapstructure:"auth" validate:"omitempty"`
		Logger         LoggerConfig    `mapstructure:"logger" validate:"omitempty"`
	}

	// HTTP server config
	ServerConfig struct {
		Address string `mapstructure:"address" validate:"required,hostname_port"`
	}

	// PostgreSQL config
	PostgresConfig struct {
		Address  string `mapstructure:"address" validate:"required,hostname_port"`
		User     string `mapstructure:"user" validate:"required"`     // TODO: check username
		Password string `mapstructure:"password" validate:"required"` // TODO: check password
		DBName   string `mapstructure:"dbname" validate:"required"`   // TODO: check dbname
		SSLMode  string `mapstructure:"sslmode" validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full"`
	}

	// Admin authentication. Not read by anything yet: user auth is deferred and
	// the API runs as a single mocked user. Kept so that enabling it later is a
	// config value and a middleware line rather than a new plumbing pass.
	AuthConfig struct {
		AdminToken string `mapstructure:"admin_token" validate:"omitempty,min=16"`
	}

	LoggerConfig struct {
		Default ComponentConfig            `mapstructure:"default"`
		Modules map[string]ComponentConfig `mapstructure:"modules"`
	}

	ComponentConfig struct {
		Level string `mapstructure:"level" validate:"required,oneof=trace debug info warn error fatal panic"`
	}
)

func (c *Config) GetLoggerConfig(module string) ComponentConfig {
	if loggerCfg, ok := c.Logger.Modules[module]; ok {
		return loggerCfg
	}
	log.Warn().Str("module", module).Msg("Not found logger config for module")
	return c.Logger.Default
}

// GetDSN builds the connection string.
//
// The credentials are escaped rather than interpolated raw: a password
// containing '@' or '/' would otherwise silently produce a DSN pointing at the
// wrong host. The result is never logged — it carries the password, and a trace
// level that someone enables during an incident is exactly when it would leak
// into a shared log. Use SafeDSN for diagnostics.
func (pc *PostgresConfig) GetDSN() string {
	sslMode := pc.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(pc.User, pc.Password),
		Host:     pc.Address,
		Path:     "/" + pc.DBName,
		RawQuery: url.Values{"sslmode": {sslMode}}.Encode(),
	}

	return dsn.String()
}

// SafeDSN is GetDSN with the password redacted, for logs and error messages.
func (pc *PostgresConfig) SafeDSN() string {
	return fmt.Sprintf("postgres://%s:***@%s/%s", pc.User, pc.Address, pc.DBName)
}

// Load load and validate config from file
func Load() *Config {
	// TODO: .env config

	// Initialize Viper
	yaml_config := viper.New()

	yaml_config.AutomaticEnv()
	yaml_config.SetEnvPrefix(EnvPrefix)
	yaml_config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// if err := yaml_config.ReadInConfig(); err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to read config")
	// 	return nil
	// }

	log.Debug().Any("path", yaml_config.GetString(("config.path"))).Msg("")
	log.Debug().Any("name", yaml_config.GetString(("config.name"))).Msg("")
	log.Debug().Any("type", yaml_config.GetString(("config.type"))).Msg("")

	// Set defaults for config initialization
	yaml_config.SetDefault("config.name", DefaultConfigName)
	yaml_config.SetDefault("config.path", DefaultConfigPath)
	yaml_config.SetDefault("config.type", ConfigType)

	// Set config search paths
	yaml_config.AddConfigPath(yaml_config.GetString(("config.path")))
	yaml_config.SetConfigType(yaml_config.GetString(("config.type")))
	yaml_config.SetConfigName(yaml_config.GetString(("config.name")))

	log.Debug().Any("keys", yaml_config.AllKeys()).Msg("all keys")

	// Read raw config
	if err := yaml_config.ReadInConfig(); err != nil {
		log.Fatal().Err(err).Msg("Failed to read config")
		return nil
	}

	// Unmarshal config
	config := new(Config)
	if err := yaml_config.Unmarshal(&config); err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config")
		return nil
	}

	// Validate config
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		log.Fatal().Err(err).Msg("Failed to validate config")
		return nil
	}

	return config
}
