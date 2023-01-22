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
	"golang.org/x/oauth2"
)

var autosr = map[string]bool{
	"#ericarei":   true,
	"#sangnope":   true,
	"#azurindayo": true,
}

var spotifyStates = map[string]struct {
	SpotifyClient *spotify.Client
}{
	// "#sangnope":   {},d. A successful call returns err == nil,
	"#azurindayo": {},
}

func StartupSpotify() {
	// setup app client
	// ctx := context.Background()
	// config := &clientcredentials.Config{
	// 	ClientID:     data.AppCfg.SpotifyID,
	// 	ClientSecret: data.AppCfg.SpotifySecret,
	// 	TokenURL:     spotifyauth.TokenURL,
	// }
	// token, err := config.Token(ctx)
	// if err != nil {
	// 	log.Fatalf("couldn't get token: %v", err)
	// }

	// httpClient := spotifyauth.New().Client(ctx, token)
	// appClient = spotify.New(httpClient)

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
		var auth oauth2.Token
		err = json.Unmarshal(b, &auth)
		if err != nil {
			log.Println("json format error for auth", c)
			continue
		}

		// auth with spotify client
		userClient := spotify.New(spotifyauth.New(
			spotifyauth.WithRedirectURL("http://localhost:51337"),
			spotifyauth.WithScopes(
				spotifyauth.ScopeUserModifyPlaybackState,
				spotifyauth.ScopeUserReadCurrentlyPlaying,
				spotifyauth.ScopeUserReadPlaybackState,
				spotifyauth.ScopeUserReadRecentlyPlayed,
			),
			spotifyauth.WithClientID(data.AppCfg.SpotifyID),
			spotifyauth.WithClientSecret(data.AppCfg.SpotifySecret),
		).Client(context.Background(), &auth))

		spotifyStates[c] = struct{ SpotifyClient *spotify.Client }{SpotifyClient: userClient}
		log.Println("Spotify client set for " + c)
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
		*msgChan <- chat("I couldn't check if the broadcaster is live ericareiCry", channel)
		return
	}
	if !live {
		*msgChan <- chat("Broadcaster is not live you silly ericareiGiggle", channel)
		return
	}
	var ok bool
	var state struct {
		SpotifyClient *spotify.Client
	}
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	brokenMsg := strings.Split(actualMessage, " ")
	for _, s := range brokenMsg {
		if strings.Contains(s, "youtube.com") || strings.Contains(s, "youtu.be") {
			// Ignore youtube song requests
			return
		}
		if strings.HasPrefix(s, "spotify:track:") || strings.HasPrefix(s, "https://open.spotify.com/track/") {
			TrackID := ""
			if strings.HasPrefix(s, "https://open.spotify.com/track/") {
				TrackID = strings.TrimPrefix(s, "https://open.spotify.com/track/")
				if strings.Contains(TrackID, "?") {
					TrackID = TrackID[:strings.Index(TrackID, "?")]
				}
			} else {
				TrackID = strings.TrimPrefix(s, "spotify:track:")
			}
			ctx := context.Background()
			result, err := state.SpotifyClient.GetTrack(ctx, spotify.ID(TrackID))
			if err != nil {
				*msgChan <- chat("Error when searching track "+err.Error(), channel)
				return
			}
			err = state.SpotifyClient.QueueSong(ctx, spotify.ID(TrackID))
			if err != nil {
				*msgChan <- chat("Error adding track to queue "+err.Error()+" sangnoSad", channel)
				return
			}
			*msgChan <- chat("Added "+result.Name+" by "+result.Artists[0].Name+" to queue", channel)
			return
		}
	}
	// text search
	ctx := context.Background()
	result, err := state.SpotifyClient.Search(ctx, actualMessage, spotify.SearchTypeTrack, spotify.Market("US"))
	if err != nil {
		*msgChan <- chat("Error when searching track"+err.Error(), channel)
		return
	}
	if len(result.Tracks.Tracks) > 0 {
		t := result.Tracks.Tracks[0].ID
		err = state.SpotifyClient.QueueSong(ctx, spotify.ID(t))
		if err != nil {
			*msgChan <- chat("Error adding track to queue "+err.Error()+" sangnoSad", channel)
			return
		}
		*msgChan <- chat("Added "+result.Tracks.Tracks[0].Name+" by "+result.Tracks.Tracks[0].Artists[0].Name+" to queue", channel)
	} else {
		*msgChan <- chat("No results found on Spotify sangnoSad", channel)
	}
}
