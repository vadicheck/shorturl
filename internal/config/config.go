package config

import "flag"

var Config struct {
	A           string
	B           string
	StoragePath string
}

func ParseFlags() {
	flag.StringVar(&Config.A, "a", "localhost:8080", "HTTP server startup address")
	flag.StringVar(&Config.B, "b", "http://localhost:8080", "the base address of the resulting shortened URL")
	flag.StringVar(&Config.StoragePath, "storage-path", "", "path to storage")
	flag.Parse()
}
