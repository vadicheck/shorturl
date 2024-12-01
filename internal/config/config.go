package config

import (
	"flag"
	"os"
)

var Config struct {
	A           string
	B           string
	StoragePath string
}

func ParseFlags() {
	Config.A = os.Getenv("SERVER_ADDRESS")
	Config.B = os.Getenv("BASE_URL")
	Config.StoragePath = os.Getenv("STORAGE_PATH")

	if Config.A == "" {
		flag.StringVar(&Config.A, "a", "localhost:8080", "HTTP server startup address")
	}

	if Config.B == "" {
		flag.StringVar(&Config.B, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	}

	if Config.StoragePath == "" {
		flag.StringVar(&Config.StoragePath, "storage-path", "", "path to storage")
	}

	flag.Parse()
}
