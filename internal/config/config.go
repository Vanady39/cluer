package config

import (
	"fmt"
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

// Get Domain Source Name from config
func (pc *PostgresConfig) GetDSN() string {
	// Escaping special characters in username/password
	log.Trace().Str("user", pc.User).Str("password", pc.Password).Str("address", pc.Address).Str("dbname", pc.DBName).Msg("check postgres config")
	// user := url.QueryEscape(pc.User)
	// password := url.QueryEscape(pc.Password)

	DSN := fmt.Sprintf("postgres://%s:%s@%s/%s", pc.User, pc.Password, pc.Address, pc.DBName)
	log.Trace().Str("DSN", DSN).Msg("trace DSN")
	return DSN
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
