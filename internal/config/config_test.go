package config

import (
	"testing"
)

// mockEnvReader creates an EnvReader from a map of key-value pairs.
func mockEnvReader(env map[string]string) EnvReader {
	return func(key string) string {
		return env[key]
	}
}

func TestLoadWithReader_Defaults(t *testing.T) {
	t.Parallel()

	cfg := LoadWithReader(mockEnvReader(map[string]string{}))

	if cfg.Minder.AuthToken != "" {
		t.Errorf("AuthToken = %q, want empty", cfg.Minder.AuthToken)
	}
	if cfg.Minder.Host != "api.stacklok.com" {
		t.Errorf("Host = %q, want %q", cfg.Minder.Host, "api.stacklok.com")
	}
	if cfg.Minder.Port != 443 {
		t.Errorf("Port = %d, want %d", cfg.Minder.Port, 443)
	}
	if cfg.Minder.Insecure != false {
		t.Errorf("Insecure = %v, want false", cfg.Minder.Insecure)
	}
	if cfg.MCP.Port != 8080 {
		t.Errorf("MCP.Port = %d, want %d", cfg.MCP.Port, 8080)
	}
	if cfg.MCP.EndpointPath != "/mcp" {
		t.Errorf("EndpointPath = %q, want %q", cfg.MCP.EndpointPath, "/mcp")
	}
}

func TestLoadWithReader_CustomValues(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"MINDER_AUTH_TOKEN":  "test-token",
		"MINDER_SERVER_HOST": "localhost",
		"MINDER_SERVER_PORT": "9090",
		"MINDER_INSECURE":    "true",
		"MCP_PORT":           "3000",
		"MCP_ENDPOINT_PATH":  "/api/mcp",
	}

	cfg := LoadWithReader(mockEnvReader(env))

	if cfg.Minder.AuthToken != "test-token" {
		t.Errorf("AuthToken = %q, want %q", cfg.Minder.AuthToken, "test-token")
	}
	if cfg.Minder.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Minder.Host, "localhost")
	}
	if cfg.Minder.Port != 9090 {
		t.Errorf("Port = %d, want %d", cfg.Minder.Port, 9090)
	}
	if cfg.Minder.Insecure != true {
		t.Errorf("Insecure = %v, want true", cfg.Minder.Insecure)
	}
	if cfg.MCP.Port != 3000 {
		t.Errorf("MCP.Port = %d, want %d", cfg.MCP.Port, 3000)
	}
	if cfg.MCP.EndpointPath != "/api/mcp" {
		t.Errorf("EndpointPath = %q, want %q", cfg.MCP.EndpointPath, "/api/mcp")
	}
}

func TestGetEnvDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		env          map[string]string
		key          string
		defaultValue string
		want         string
	}{
		{
			name:         "returns env value when set",
			env:          map[string]string{"KEY": "value"},
			key:          "KEY",
			defaultValue: "default",
			want:         "value",
		},
		{
			name:         "returns default when key not present",
			env:          map[string]string{},
			key:          "KEY",
			defaultValue: "default",
			want:         "default",
		},
		{
			name:         "returns default when value is empty",
			env:          map[string]string{"KEY": ""},
			key:          "KEY",
			defaultValue: "default",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getEnvDefault(mockEnvReader(tt.env), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		env          map[string]string
		key          string
		defaultValue int
		want         int
	}{
		{
			name:         "parses valid integer",
			env:          map[string]string{"PORT": "8080"},
			key:          "PORT",
			defaultValue: 443,
			want:         8080,
		},
		{
			name:         "returns default for missing key",
			env:          map[string]string{},
			key:          "PORT",
			defaultValue: 443,
			want:         443,
		},
		{
			name:         "returns default for invalid integer",
			env:          map[string]string{"PORT": "not-a-number"},
			key:          "PORT",
			defaultValue: 443,
			want:         443,
		},
		{
			name:         "returns default for empty value",
			env:          map[string]string{"PORT": ""},
			key:          "PORT",
			defaultValue: 443,
			want:         443,
		},
		{
			name:         "handles negative integers",
			env:          map[string]string{"OFFSET": "-10"},
			key:          "OFFSET",
			defaultValue: 0,
			want:         -10,
		},
		{
			name:         "handles zero",
			env:          map[string]string{"COUNT": "0"},
			key:          "COUNT",
			defaultValue: 5,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getEnvInt(mockEnvReader(tt.env), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		env          map[string]string
		key          string
		defaultValue bool
		want         bool
	}{
		{
			name:         "parses 'true'",
			env:          map[string]string{"FLAG": "true"},
			key:          "FLAG",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "parses 'false'",
			env:          map[string]string{"FLAG": "false"},
			key:          "FLAG",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "parses '1' as true",
			env:          map[string]string{"FLAG": "1"},
			key:          "FLAG",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "parses '0' as false",
			env:          map[string]string{"FLAG": "0"},
			key:          "FLAG",
			defaultValue: true,
			want:         false,
		},
		{
			name:         "returns default for missing key",
			env:          map[string]string{},
			key:          "FLAG",
			defaultValue: true,
			want:         true,
		},
		{
			name:         "returns default for invalid bool",
			env:          map[string]string{"FLAG": "yes"},
			key:          "FLAG",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "returns default for empty value",
			env:          map[string]string{"FLAG": ""},
			key:          "FLAG",
			defaultValue: true,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := getEnvBool(mockEnvReader(tt.env), tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("getEnvBool() = %v, want %v", got, tt.want)
			}
		})
	}
}
