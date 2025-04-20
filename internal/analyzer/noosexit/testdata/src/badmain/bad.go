package main

import "os"

func main() {
	os.Exit(1) // want "use of os.Exit in main.main is forbidden"
}
