package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds application configuration
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	Scanner  ScannerConfig  `mapstructure:"scanner"`
	Server   ServerConfig   `mapstructure:"server"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	File   string `mapstructure:"file"`
}

// ScannerConfig holds scanner configuration
type ScannerConfig struct {
	DefaultTimeout int               `mapstructure:"default_timeout"`
	MaxThreads     int               `mapstructure:"max_threads"`
	DefaultPorts   string            `mapstructure:"default_ports"`
	Presets        map[string]Preset `mapstructure:"presets"`
}

// Preset holds scanner preset configuration
type Preset struct {
	Scanner   string `mapstructure:"scanner"`
	Ports     string `mapstructure:"ports"`
	Arguments string `mapstructure:"arguments"`
	Timing    string `mapstructure:"timing"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "netrecon")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "text")
	viper.SetDefault("logging.file", "")

	viper.SetDefault("scanner.default_timeout", 300)
	viper.SetDefault("scanner.max_threads", 1000)
	viper.SetDefault("scanner.default_ports", "1-1000")

	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)

	// Set environment variable prefix
	viper.SetEnvPrefix("NETRECON")
	viper.AutomaticEnv()

	// Set configuration file name and paths
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath("$HOME/.netrecon")
		viper.AddConfigPath("/etc/netrecon")
	}

	// Read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configDir := filepath.Join(homeDir, ".netrecon")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	viper.Set("database", config.Database)
	viper.Set("logging", config.Logging)
	viper.Set("scanner", config.Scanner)
	viper.Set("server", config.Server)

	return viper.WriteConfigAs(configPath)
}
