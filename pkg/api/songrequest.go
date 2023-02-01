package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

/*
Adding channel for autosr:
1) Add channel name to autosr map
2) If using spotify, get credentials using cmd/spotifyoauth/main.go, and add channel name spotifyStates
*/

var autosr = map[string]bool{
	"#ericarei":   true,
	"#sangnope":   true,
	"#azurindayo": true,
}

var spotifyStates = map[string]struct {
	SpotifyClient *spotify.Client
	LastSkip      time.Time
}{
	"#sangnope": {
		SpotifyClient: nil,
		LastSkip:      time.Now(),
	},
	"#azurindayo": {
		SpotifyClient: nil,
		LastSkip:      time.Now(),
	},
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

		spotifyStates[c] = struct {
			SpotifyClient *spotify.Client
			LastSkip      time.Time
		}{
			SpotifyClient: userClient,
			LastSkip:      time.Now(),
		}
		log.Println("Spotify client set for " + c)
	}
}

func processSongRequestNightBot(msgChan *chan string, channel string, actualMessage string) {
	if autosr[channel] {
		*msgChan <- chat("!sr "+actualMessage, channel)
	}
}

func commandSkipSongSpotify(channel string, user string, acutalMessage string) {
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
	if _, ok = spotifyStates[channel]; !ok {
		*msgChan <- chat("Uh... AzuriBot broken sangnoDead ...", channel)
		log.Println("Error: spotify state dne", channel)
		return
	}
	if spotifyStates[channel].SpotifyClient == nil {
		*msgChan <- chat("Uh... AzuriBot broken sangnoDead ...", channel)
		log.Println("Error: Calling nil state.SpotifyClient fopr channel", channel)
		return
	}
	now := time.Now()
	if spotifyStates[channel].LastSkip.Add(time.Second * 10).After(now) {
		*msgChan <- chat("Don't skip song too quickly! ericareiPout", channel)
		return
	}
	newState := spotifyStates[channel]
	newState.LastSkip = now
	spotifyStates[channel] = newState

	ctx := context.Background()
	err = spotifyStates[channel].SpotifyClient.Next(ctx)
	if err != nil {
		*msgChan <- chat("Failed to add song sangnoSad "+err.Error(), channel)
		return
	}
	*msgChan <- chat("Skipped song sangnoWave", channel)
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
		LastSkip      time.Time
	}
	if state, ok = spotifyStates[channel]; !ok {
		return
	}
	if state.SpotifyClient == nil {
		// log.Println("Error: Calling nil state.SpotifyClient fopr channel", channel)
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

		queue, err := state.SpotifyClient.GetQueue(context.Background())
		if err != nil {
			*msgChan <- chat("Error Couldn't check if your song was already queued "+err.Error()+" sangnoSad", channel)
			return
		}
		if queue.CurrentlyPlaying.ID == t {
			*msgChan <- chat("It's the currently playing song you silly ericareiGiggle", channel)
			return
		}
		for _, v := range queue.Items {
			if v.ID == t {
				*msgChan <- chat("That song is already queued you silly ericareiGiggle", channel)
				return
			}
		}

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
