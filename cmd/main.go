package main

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/slazurin/twitch-ban-negate/pkg/api"
	"github.com/slazurin/twitch-ban-negate/pkg/data"
)

func main() {
	err := cleanenv.ReadConfig(".env", &data.AppCfg)
	if err != nil {
		err = cleanenv.ReadEnv(&data.AppCfg)
		if err != nil {
			log.Println("Failed to load env.")
			os.Exit(1)
		}
	}

	var exitCh = make(chan struct{}, 1)

	go api.Run(&exitCh)
	<-exitCh
}
