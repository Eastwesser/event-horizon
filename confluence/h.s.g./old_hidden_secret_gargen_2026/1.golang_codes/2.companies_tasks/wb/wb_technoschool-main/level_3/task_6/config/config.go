package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	DSN string `mapstructure:"dsn"`
}

func Load() (*Config, error) {
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Set defaults
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("database.dsn", os.Getenv("DATABASE_DSN"))

	// Read from env
	viper.BindEnv("database.dsn", "DATABASE_DSN")

	// Try to read config file, but don't fail if it doesn't exist
	_ = viper.ReadInConfig()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}

