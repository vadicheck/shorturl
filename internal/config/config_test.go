package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestParseFlags_DefaultValues(t *testing.T) {
	// Сбросим переменные окружения и аргументы командной строки
	os.Clearenv()
	resetArgs()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	cfg := Config

	if cfg.AppEnv != "prod" {
		t.Errorf("expected AppEnv to be 'prod', got '%s'", cfg.AppEnv)
	}
	if cfg.ServerAddress != "localhost:8080" {
		t.Errorf("expected ServerAddress to be 'localhost:8080', got '%s'", cfg.ServerAddress)
	}
	if cfg.JwtTokenExpire != 24*time.Hour {
		t.Errorf("expected JwtTokenExpire to be 24h, got '%s'", cfg.JwtTokenExpire)
	}
	if cfg.SecureCookieHashKey != "very-secret" {
		t.Errorf("expected default SecureCookieHashKey, got '%s'", cfg.SecureCookieHashKey)
	}
	if cfg.SecureCookieBlockKey != "alotsecretalotsecretalotsecretgr" {
		t.Errorf("expected default SecureCookieBlockKey, got '%s'", cfg.SecureCookieBlockKey)
	}
}

func TestParseFlags_WithEnvOverrides(t *testing.T) {
	os.Setenv("APP_ENV", "dev")
	os.Setenv("SERVER_ADDRESS", "127.0.0.1:9999")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("JWT_SECRET", "env-secret")
	os.Setenv("SECURE_COOKIE_HASH_KEY", "env-hash")
	os.Setenv("SECURE_COOKIE_BLOCK_KEY", "env-block")

	resetArgs()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	cfg := Config

	if cfg.AppEnv != "dev" {
		t.Errorf("expected AppEnv to be 'dev', got '%s'", cfg.AppEnv)
	}
	if cfg.ServerAddress != "127.0.0.1:9999" {
		t.Errorf("expected ServerAddress to be '127.0.0.1:9999', got '%s'", cfg.ServerAddress)
	}
	if !cfg.EnableHTTPS {
		t.Errorf("expected EnableHTTPS to be true")
	}
	if cfg.JwtSecret != "env-secret" {
		t.Errorf("expected JwtSecret to be 'env-secret', got '%s'", cfg.JwtSecret)
	}
	if cfg.SecureCookieHashKey != "env-hash" {
		t.Errorf("expected SecureCookieHashKey to be 'env-hash', got '%s'", cfg.SecureCookieHashKey)
	}
	if cfg.SecureCookieBlockKey != "env-block" {
		t.Errorf("expected SecureCookieBlockKey to be 'env-block', got '%s'", cfg.SecureCookieBlockKey)
	}
}

func TestParseFlags_JSONConfig(t *testing.T) {
	tmpFile := createTempJSONConfig(t, `{
		"app_env": "json-env",
		"server_address": "json:9000",
		"base_url": "http://json.local",
		"database_dsn": "postgres://user:pass@localhost/db"
	}`)

	os.Setenv("CONFIG", tmpFile)
	resetArgs()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	ParseFlags()

	cfg := Config

	if cfg.AppEnv != "dev" {
		t.Errorf("expected AppEnv from JSON, got '%s'", cfg.AppEnv)
	}
	if cfg.ServerAddress != "127.0.0.1:9999" {
		t.Errorf("expected ServerAddress from JSON, got '%s'", cfg.ServerAddress)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("expected BaseURL from JSON, got '%s'", cfg.BaseURL)
	}
	if cfg.DatabaseDsn != "postgres://user:pass@localhost/db" {
		t.Errorf("expected DatabaseDsn from JSON, got '%s'", cfg.DatabaseDsn)
	}
}

// resetArgs resets os.Args to default
func resetArgs() {
	os.Args = []string{"cmd"}
}

// createTempJSONConfig creates a temporary JSON file and returns its path.
func createTempJSONConfig(t *testing.T, content string) string {
	t.Helper()

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}

	return tmpFile.Name()
}

func Test_isZero(t *testing.T) {
	type testStruct struct {
		Field1 string
		Field2 bool
		Field3 int64
		Field4 time.Duration
	}

	tests := []struct {
		name     string
		fieldVal any
		expected bool
	}{
		{"empty string", "", true},
		{"non-empty string", "abc", false},
		{"false bool", false, true},
		{"true bool", true, false},
		{"zero int64", int64(0), true},
		{"non-zero int64", int64(42), false},
		{"zero duration", time.Duration(0), true},
		{"non-zero duration", time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.fieldVal)
			got := isZero(val)
			if got != tt.expected {
				t.Errorf("IsZero(%v) = %v; want %v", tt.fieldVal, got, tt.expected)
			}
		})
	}
}
