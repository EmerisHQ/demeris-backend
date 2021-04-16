package main

import (
	"log"

	"github.com/allinbits/navigator-backend/router"
)

func main() {
	r := router.New()
	log.Fatal(r.Run(":8080"))
}
