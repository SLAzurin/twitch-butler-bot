package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/slazurin/twitch-butler-bot/pkg/api"
	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"github.com/zmb3/spotify/v2"
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
	apidb.ManualInit()

	api.StartupSpotify()

	flag.Parse()

	channel := flag.Arg(0)
	if !strings.HasPrefix(channel, "#") {
		channel = "#" + channel
	}
	link := flag.Arg(1)

	state := api.GetSpotifyState(channel)
	if state.SpotifyClient == nil {
		log.Fatalln("Spotify client is nil")
	}

	log.Println(state.SpotifyClient.QueueSong(context.Background(), spotify.ID(utils.LinkToID(link))))
}
