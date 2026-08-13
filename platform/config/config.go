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
		OIDCConfig     *OIDCConfig     `mapstructure:"oidc" validate:"omitempty"`
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

	OIDCConfig struct {
		Issuer      string   `mapstructure:"issuer" validate:"required,url"`
		ClientID    string   `mapstructure:"client_id" validate:"required"`
		AdminGroups []string `mapstructure:"admin_groups" validate:"omitempty,min=1,dive,required"`
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
