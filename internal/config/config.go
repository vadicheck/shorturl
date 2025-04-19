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
// - TLSCertPath: cert path.
// - TLSKeyPath: key path.
//
// The configuration can be specified via command-line flags or environment variables, with the
// environment variables taking precedence over flag values.
package config

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"
)

const defaultJwtHours = 24

// Config holds the configuration values for the application.
var Config struct {
	AppEnv               string
	ServerAddress        string
	BaseURL              string
	DatabaseDsn          string
	FileStoragePath      string
	JwtSecret            string
	JwtTokenExpire       time.Duration
	SecureCookieHashKey  string
	SecureCookieBlockKey string
	SecureCookieExpire   time.Duration
	EnableHTTPS          bool
	TLSCertPath          string
	TLSKeyPath           string
}

// ParseFlags parses command-line flags and environment variables to populate the configuration values.
// Default values are set in the code, but they can be overridden by flags or environment variables.
func ParseFlags() {
	flag.StringVar(&Config.AppEnv, "e", "prod", "environment")
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.DatabaseDsn, "d", "", "database DSN")
	flag.StringVar(&Config.FileStoragePath, "f", "./storage/filestorage.txt", "path to file storage")
	flag.BoolVar(&Config.EnableHTTPS, "s", false, "enable HTTPS")
	flag.StringVar(&Config.TLSCertPath, "c", "certs/localhost.pem", "path to TLS cert")
	flag.StringVar(&Config.TLSKeyPath, "k", "certs/localhost-key.pem", "path to TLS key")

	flag.Parse()

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
