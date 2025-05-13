// Package config provides functions for parsing the application's configuration.
//
// The configuration is loaded from command-line flags and environment variables.
// Default values are set in the code but can be overridden by the user at runtime.
//
// The following configuration values can be set through flags or environment variables:
// - AppEnv: The environment in which the application is running (e.g., "prod", "dev").
// - ServerAddress: The address on which the HTTP server will listen (e.g., "localhost:8080").
// - BaseURL: The base address of the resulting shortened URL (e.g., "http://localhost:8080").
// - DatabaseDsn: The Data Source Name (DSN) for connecting to the database.
// - FileStoragePath: The path to the file used for storing data in file storage.
// - JwtSecret: The secret key used to sign JWT tokens.
// - JwtTokenExpire: The duration for which JWT tokens are valid.
// - SecureCookieHashKey: The key used for securing cookies in hashing.
// - SecureCookieBlockKey: The key used for securing cookies in encryption.
// - SecureCookieExpire: The duration for which cookies are valid.
// - EnableHTTPS: Enable HTTPS on server.
// - TLSCertPath: Cert path.
// - TLSKeyPath: Key path.
// - JSONConfig: Path to JSON config.
// - TrustedSubnet: Allowed subnet for statistics.
//
// The configuration can be specified via command-line flags or environment variables, with the
// environment variables taking precedence over flag values.
package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

const defaultJwtHours = 24

// CfgStruct holds the configuration values for the application.
type CfgStruct struct {
	AppEnv               string `json:"app_env"`
	ServerAddress        string `json:"server_address"`
	BaseURL              string `json:"base_url"`
	DatabaseDsn          string `json:"database_dsn"`
	FileStoragePath      string `json:"file_storage_path"`
	JwtSecret            string `json:"jwt_secret"`
	JwtTokenExpire       time.Duration
	SecureCookieHashKey  string `json:"secure_cookie_hash_key"`
	SecureCookieBlockKey string `json:"secure_cookie_block_key"`
	SecureCookieExpire   time.Duration
	EnableHTTPS          bool   `json:"enable_https"`
	TLSCertPath          string `json:"tls_cert_path"`
	TLSKeyPath           string `json:"tls_key_path"`
	TrustedSubnet        string `json:"trusted_subnet"`
	JSONConfig           string
}

// Config is the global instance of CfgStruct used by the application.
//
// It holds all runtime configuration values and is typically initialized
// during startup via a combination of environment variables, command-line
// arguments, and JSON configuration files.
var Config CfgStruct

// ParseFlags parses command-line flags and environment variables to populate the configuration values.
// Default values are set in the code, but they can be overridden by flags or environment variables.
func ParseFlags() {
	flag.StringVar(&Config.AppEnv, "e", "prod", "environment")
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.DatabaseDsn, "d", "", "database DSN")
	flag.StringVar(&Config.FileStoragePath, "f", "./storage/filestorage.txt", "path to file storage")
	flag.BoolVar(&Config.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&Config.TLSCertPath, "p", "certs/localhost.pem", "path to TLS cert")
	flag.StringVar(&Config.TLSKeyPath, "k", "certs/localhost-key.pem", "path to TLS key")
	flag.StringVar(&Config.JSONConfig, "c", "", "path to json config")
	flag.StringVar(&Config.TrustedSubnet, "t", "", "trusted subnet")

	flag.Parse()

	if envJSONConfig := os.Getenv("CONFIG"); envJSONConfig != "" {
		Config.JSONConfig = envJSONConfig
	}

	if Config.JSONConfig != "" {
		parseJSONConfig()
	}

	Config.JwtTokenExpire = time.Hour * defaultJwtHours
	Config.SecureCookieExpire = time.Hour * defaultJwtHours

	if appEnv := os.Getenv("APP_ENV"); appEnv != "" {
		Config.AppEnv = appEnv
	}

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		Config.ServerAddress = serverAddress
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Config.BaseURL = baseURL
	}

	if databaseDsn := os.Getenv("DATABASE_DSN"); databaseDsn != "" {
		Config.DatabaseDsn = databaseDsn
	}

	if fileStoragePath := os.Getenv("FILE_STORAGE_PATH"); fileStoragePath != "" {
		Config.FileStoragePath = fileStoragePath
	}

	if enableHTTPS := os.Getenv("ENABLE_HTTPS"); enableHTTPS != "" {
		parsed, err := strconv.ParseBool(enableHTTPS)
		if err != nil {
			log.Fatalf("invalid ENABLE_HTTPS value: %v", err)
		}
		Config.EnableHTTPS = parsed
	}

	if TLSCertPath := os.Getenv("TLS_CERT_PATH"); TLSCertPath != "" {
		Config.TLSCertPath = TLSCertPath
	}

	if TLSKeyPath := os.Getenv("TLS_KEY_PATH"); TLSKeyPath != "" {
		Config.TLSKeyPath = TLSKeyPath
	}

	if TrustedSubnet := os.Getenv("TRUSTED_SUBNET"); TrustedSubnet != "" {
		Config.TrustedSubnet = TrustedSubnet
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		Config.JwtSecret = jwtSecret
	} else {
		Config.JwtSecret = "secretkey"
	}

	if secureCookieHashKey := os.Getenv("SECURE_COOKIE_HASH_KEY"); secureCookieHashKey != "" {
		Config.SecureCookieHashKey = secureCookieHashKey
	} else {
		Config.SecureCookieHashKey = "very-secret"
	}

	if secureCookieBlockKey := os.Getenv("SECURE_COOKIE_BLOCK_KEY"); secureCookieBlockKey != "" {
		Config.SecureCookieBlockKey = secureCookieBlockKey
	} else {
		Config.SecureCookieBlockKey = "alotsecretalotsecretalotsecretgr"
	}
}

// parseJSONConfig reads a JSON configuration file from the path specified
// in Config.JSONConfig, parses it into a CfgStruct, and updates the global
// Config variable by copying all fields that are not set (zero values) in Config
// from the parsed config file.
func parseJSONConfig() {
	cfgJSON := &CfgStruct{}

	file, err := os.ReadFile(Config.JSONConfig)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(file, &cfgJSON)
	if err != nil {
		log.Panic(err)
	}

	copyMissingFields(&Config, *cfgJSON)
}

// copyMissingFields copies all non-zero fields from the 'from' CfgStruct
// into the 'to' CfgStruct pointer, but only for fields that are currently
// zero-valued in 'to'.
//
// Fields are compared and copied using reflection.
func copyMissingFields(to *CfgStruct, from CfgStruct) {
	toVal := reflect.ValueOf(to).Elem()
	fromVal := reflect.ValueOf(from)

	for i := 0; i < toVal.NumField(); i++ {
		toField := toVal.Field(i)
		fromField := fromVal.Field(i)

		if isZero(toField) && !isZero(fromField) {
			toField.Set(fromField)
		}
	}
}

// isZero returns true if the given reflect.Value is considered zero (empty).
//
// Supported types: string, bool, int, int64. For other types, reflect.IsZero is used.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int64:
		return v.Int() == 0
	default:
		return v.IsZero()
	}
}
