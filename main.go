package main

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/router"
)

func main() {
	config.InitConfig()
	config.InitDatabase()
	router := router.SetupRouter()
	router.Run()
}
