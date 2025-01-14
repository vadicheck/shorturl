package config

import (
	"flag"
	"os"
)

var Config struct {
	ServerAddress   string
	BaseURL         string
	DatabaseDsn     string
	FileStoragePath string
}

func ParseFlags() {
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.DatabaseDsn, "d", "", "database DSN")
	flag.StringVar(&Config.FileStoragePath, "f", "./storage/filestorage.txt", "path to file storage")

	flag.Parse()

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
}
