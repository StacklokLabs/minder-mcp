// Package config provides configuration loading for the MCP server.
package config

import (
	"os"
	"strconv"
)

// EnvReader is a function type for reading environment variables.
type EnvReader func(key string) string

// OSEnvReader is the default EnvReader that uses os.Getenv.
func OSEnvReader(key string) string {
	return os.Getenv(key)
}

// Config holds all configuration for the MCP server.
type Config struct {
	Minder MinderConfig
	MCP    MCPConfig
}

// MinderConfig holds Minder-specific configuration.
type MinderConfig struct {
	AuthToken string
	Host      string
	Port      int
	Insecure  bool
}

// MCPConfig holds MCP server configuration.
type MCPConfig struct {
	Port         int
	EndpointPath string
}

// Load reads configuration from environment variables using the default OS reader.
func Load() *Config {
	return LoadWithReader(OSEnvReader)
}

// LoadWithReader reads configuration using the provided EnvReader.
func LoadWithReader(getEnv EnvReader) *Config {
	return &Config{
		Minder: MinderConfig{
			AuthToken: getEnvDefault(getEnv, "MINDER_AUTH_TOKEN", ""),
			Host:      getEnvDefault(getEnv, "MINDER_SERVER_HOST", "api.stacklok.com"),
			Port:      getEnvInt(getEnv, "MINDER_SERVER_PORT", 443),
			Insecure:  getEnvBool(getEnv, "MINDER_INSECURE", false),
		},
		MCP: MCPConfig{
			Port:         getEnvInt(getEnv, "MCP_PORT", 8080),
			EndpointPath: getEnvDefault(getEnv, "MCP_ENDPOINT_PATH", "/mcp"),
		},
	}
}

func getEnvDefault(getEnv EnvReader, key, defaultValue string) string {
	if value := getEnv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(getEnv EnvReader, key string, defaultValue int) int {
	if value := getEnv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(getEnv EnvReader, key string, defaultValue bool) bool {
	if value := getEnv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
