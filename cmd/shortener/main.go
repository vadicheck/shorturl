package main

import (
	"log"

	"github.com/vadicheck/shorturl/internal/app"
)

func main() {
	httpApp := app.New()

	if err := httpApp.Run(); err != nil {
		log.Panic(err)
	}
}
