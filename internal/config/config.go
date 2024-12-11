package config

import (
	"flag"
	"os"
)

var Config struct {
	ServerAddress string
	BaseURL       string
	StoragePath   string
}

func ParseFlags() {
	flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.StoragePath, "storage-path", "", "path to storage")

	flag.Parse()

	if serverAddress := os.Getenv("SERVER_ADDRESS"); serverAddress != "" {
		Config.ServerAddress = serverAddress
	}

	if baseURL := os.Getenv("BASE_URL"); baseURL != "" {
		Config.BaseURL = baseURL
	}

	if storagePath := os.Getenv("STORAGE_PATH"); storagePath != "" {
		Config.StoragePath = storagePath
	}
}
