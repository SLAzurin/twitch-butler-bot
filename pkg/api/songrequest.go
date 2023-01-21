package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

var autosr = map[string]bool{
	"#ericarei":   true,
	"#sangnope":   true,
	"#azurindayo": true,
}
var appClient *spotify.Client

var spotifyStates = map[string]struct {
	SpotifyClient *spotify.Client
	SpotifyCtx    *context.Context
	SpotifyAuth   data.SpotifyAuth
}{
	// "#sangnope":   {},d. A successful call returns err == nil,
	"#azurindayo": {},
}

func StartupSpotify() {
	// setup app client
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     data.AppCfg.SpotifyID,
		ClientSecret: data.AppCfg.SpotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	appClient = spotify.New(httpClient)

	// Read auth and get auth
	for c := range spotifyStates {
		jsonFile, err := os.Open("tokens/" + c + ".json")
		if err != nil {
			log.Println("No auth for", c)
			continue
		}
		defer jsonFile.Close()
		b, err := io.ReadAll(jsonFile)
		if err != nil {
			log.Println("Cannot read auth json", c)
			continue
		}
		var auth data.SpotifyAuth
		err = json.Unmarshal(b, &auth)
		if err != nil {
			log.Println("json format error for auth", c)
			continue
		}
		// auth with spotify client
		appClient.CurrentUser()
	}
}

func processSongRequestNightBot(msgChan *chan string, channel string, actualMessage string) {
	if autosr[channel] {
		*msgChan <- chat("!sr "+actualMessage, channel)
	}
}

func processSongRequestSpotify(msgChan *chan string, channel string, actualMessage string) {
	if !autosr[channel] {
		return
	}
	live, err := utils.ChannelIsLive(strings.Trim(channel, "#"))
	if err != nil {
		log.Println("Cannot check live channel", channel, err)
	}
	if !live {
		return
	}
	var ok bool
	var state struct {
		SpotifyClient *spotify.Client
		SpotifyCtx    *context.Context
		SpotifyAuth   data.SpotifyAuth
	}
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	brokenMsg := strings.Split(actualMessage, " ")
	for _, s := range brokenMsg {
		if strings.HasPrefix(s, "spotify:track:") || strings.HasPrefix(s, "https://open.spotify.com/track/") {
			// TODO: implement this
			*msgChan <- chat("not yet implemented", channel)
		} else {
			// text search
			ctx := context.Background()
			result, err := state.SpotifyClient.Search(ctx, actualMessage, spotify.SearchTypeTrack, spotify.Market("US"))
			if err != nil {
				log.Println("error when searching track", actualMessage, err)
			}
			if result.Tracks.Total > 0 {
				t := result.Tracks.Next
				err = state.SpotifyClient.QueueSong(ctx, spotify.ID(t))
				if err != nil {
					log.Println("error adding track to queue")
				}
				*msgChan <- chat("Added a song to queue", channel)
			} else {
				*msgChan <- chat("No results found on Spotify", channel)
			}
		}
	}
}
