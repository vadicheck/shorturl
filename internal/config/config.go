package config

import (
	"flag"
	"os"
)

var Config struct {
	ServerAddress string
	BaseUrl       string
	StoragePath   string
}

func ParseFlags() {
	Config.ServerAddress = os.Getenv("SERVER_ADDRESS")
	Config.BaseUrl = os.Getenv("BASE_URL")
	Config.StoragePath = os.Getenv("STORAGE_PATH")

	if Config.ServerAddress == "" {
		flag.StringVar(&Config.ServerAddress, "a", "localhost:8080", "HTTP server startup address")
	}

	if Config.BaseUrl == "" {
		flag.StringVar(&Config.BaseUrl, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	}

	if Config.StoragePath == "" {
		flag.StringVar(&Config.StoragePath, "storage-path", "", "path to storage")
	}

	flag.Parse()
}
