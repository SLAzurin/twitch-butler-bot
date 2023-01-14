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
		log.Println(".env failed to parse")
		os.Exit(1)
	}

	var exitCh = make(chan struct{}, 1)

	go api.Run(&exitCh)
	<-exitCh
}
