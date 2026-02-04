// Package config provides configuration loading for the MCP server.
package config

import (
	"os"
	"strconv"
)

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

// Load reads configuration from environment variables.
func Load() *Config {
	return &Config{
		Minder: MinderConfig{
			AuthToken: getEnv("MINDER_AUTH_TOKEN", ""),
			Host:      getEnv("MINDER_SERVER_HOST", "api.stacklok.com"),
			Port:      getEnvInt("MINDER_SERVER_PORT", 443),
			Insecure:  getEnvBool("MINDER_INSECURE", false),
		},
		MCP: MCPConfig{
			Port:         getEnvInt("MCP_PORT", 8080),
			EndpointPath: getEnv("MCP_ENDPOINT_PATH", "/mcp"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
