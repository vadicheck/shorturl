package config

import (
	"flag"
	"os"
	"time"
)

const defaultJwtHours = 24

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
}

func ParseFlags() {
	flag.StringVar(&Config.AppEnv, "e", "prod", "environment")
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.DatabaseDsn, "d", "", "database DSN")
	flag.StringVar(&Config.FileStoragePath, "f", "./storage/filestorage.txt", "path to file storage")

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
