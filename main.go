package main

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/router"
	"github.com/antidote-kt/SSE_Library-back/utils"
)

func main() {
	config.InitConfig()
	config.InitDatabase()
	config.InitRedis()
	config.InitEmail()
	go utils.WSManager.Start()
	router := router.SetupRouter()
	router.Run()
}
