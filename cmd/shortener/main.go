package main

import (
	"github.com/vadicheck/shorturl/internal/app"
	"log"
)

func main() {
	httpApp := app.New()

	if err := httpApp.Run(); err != nil {
		log.Panic(err)
	}
}
