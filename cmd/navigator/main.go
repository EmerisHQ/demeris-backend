package main

import (
	"log"

	"github.com/allinbits/navigator-backend/config"
	"github.com/allinbits/navigator-backend/database"
	"github.com/allinbits/navigator-backend/router"
)

func main() {

	cfg, err := config.Read()

	if err != nil {
		panic(err)
	}

	if database.Init(cfg) != nil {
		panic(err)
	}

	defer database.Close()

	r := router.New()
	log.Fatal(r.Run(":8080"))
}
