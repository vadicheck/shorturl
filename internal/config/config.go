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
	Config.ServerAddress = os.Getenv("SERVER_ADDRESS")
	Config.BaseURL = os.Getenv("BASE_URL")
	Config.StoragePath = os.Getenv("STORAGE_PATH")

	if Config.ServerAddress == "" {
		flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	}

	if Config.BaseURL == "" {
		flag.StringVar(&Config.BaseURL, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	}

	if Config.StoragePath == "" {
		flag.StringVar(&Config.StoragePath, "storage-path", "", "path to storage")
	}

	flag.Parse()
}
