package api

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

/*
Adding channel for autosr: (nightbot just link rewardsmap)
1) If using spotify, get credentials using cmd/spotifyoauth/main.go, and add them to tokens folder
*/

var spotifyStates = map[string]struct {
	SpotifyClient *spotify.Client
	LastSkip      time.Time
}{}

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
	files, err := os.ReadDir("tokens")
	if err != nil {
		log.Fatalln(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		c := file.Name()
		jsonFile, err := os.Open("tokens/" + c)
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

		spotifyStates[strings.TrimSuffix(c, ".json")] = struct {
			SpotifyClient *spotify.Client
			LastSkip      time.Time
		}{
			SpotifyClient: userClient,
			LastSkip:      time.Now(),
		}
		log.Println("Spotify client set for " + c)
	}
}

func processSongRequestNightBot(msgChan *chan string, channel string, permissionLevel int, brokenMessage []string) {
	val, err := apidb.RedisDB.Get(context.Background(), channel+"_!autosr").Result()
	if err != nil && err.Error() != "redis: nil" {
		*msgChan <- chat("I couldn't check if automatic song requests were enabled ericareiCry", channel)
		return
	}
	if val == "false" {
		return
	}
	*msgChan <- chat("!sr "+strings.Join(brokenMessage, " "), channel)

}

func commandSkipSongSpotify(channel string, user string, permissionLevel int, brokenMessage []string) {
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

func processSongRequestSpotify(msgChan *chan string, channel string, permissionLevel int, brokenMsg []string) {
	val, err := apidb.RedisDB.Get(context.Background(), channel+"_!autosr").Result()
	if err != nil && err.Error() != "redis: nil" {
		*msgChan <- chat("I couldn't check if automatic song requests were enabled sangnoSad", channel)
		return
	}
	if val == "false" {
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
	TrackID := ""
	for _, s := range brokenMsg {
		if strings.Contains(s, "youtube.com") || strings.Contains(s, "youtu.be") {
			// Ignore youtube song requests
			*msgChan <- chat("!spotify", channel)
			return
		}
		if strings.HasPrefix(s, "spotify:track:") || strings.HasPrefix(s, "https://open.spotify.com/track/") {
			if strings.HasPrefix(s, "https://open.spotify.com/track/") {
				TrackID = strings.TrimPrefix(s, "https://open.spotify.com/track/")
				if strings.Contains(TrackID, "?") {
					TrackID = TrackID[:strings.Index(TrackID, "?")]
				}
			} else {
				TrackID = strings.TrimPrefix(s, "spotify:track:")
			}
			break
		}
	}
	// text search
	if TrackID == "" {
		ctx := context.Background()
		result, err := state.SpotifyClient.Search(ctx, strings.Join(brokenMsg[1:], " "), spotify.SearchTypeTrack, spotify.Market("US"))
		if err != nil {
			*msgChan <- chat("Error when searching track "+err.Error()+" sangnoSad", channel)
			return
		}
		if len(result.Tracks.Tracks) > 0 {
			TrackID = result.Tracks.Tracks[0].ID.String()
		} else {
			*msgChan <- chat("No results found on Spotify sangnoSad", channel)
			return
		}
	}
	result, err := state.SpotifyClient.GetTrack(context.Background(), spotify.ID(TrackID))
	if err != nil {
		*msgChan <- chat("No results found on Spotify sangnoSad", channel)
		return
	}
	// Check if song is in queue already
	queue, err := state.SpotifyClient.GetQueue(context.Background())
	if err != nil {
		*msgChan <- chat("Error: Couldn't check if your song was already queued "+err.Error()+" sangnoSad", channel)
		return
	}
	if queue.CurrentlyPlaying.ID.String() == TrackID {
		*msgChan <- chat("It's the currently playing song you silly ericareiGiggle", channel)
		return
	}
	for _, v := range queue.Items {
		if v.ID.String() == TrackID {
			*msgChan <- chat("That song is already queued you silly ericareiGiggle", channel)
			return
		}
	}

	err = state.SpotifyClient.QueueSong(context.Background(), spotify.ID(TrackID))
	if err != nil {
		*msgChan <- chat("Error adding track to queue "+err.Error()+" sangnoSad", channel)
		return
	}
	*msgChan <- chat("Added "+result.Name+" by "+result.Artists[0].Name+" to queue", channel)
}

func GetSpotifyState(channel string) struct {
	SpotifyClient *spotify.Client
	LastSkip      time.Time
} {
	if v, ok := spotifyStates[channel]; ok {
		return v
	}
	return struct {
		SpotifyClient *spotify.Client
		LastSkip      time.Time
	}{}
}
