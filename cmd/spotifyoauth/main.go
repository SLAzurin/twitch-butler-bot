package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var state = "state-azurindayo"
var auth = spotifyauth.New(
	spotifyauth.WithRedirectURL("http://localhost:51337"),
	spotifyauth.WithScopes(
		spotifyauth.ScopeUserModifyPlaybackState,
		spotifyauth.ScopeUserReadCurrentlyPlaying,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserReadRecentlyPlayed,
	),
	spotifyauth.WithClientID(data.AppCfg.SpotifyID),
	spotifyauth.WithClientSecret(data.AppCfg.SpotifySecret),
)
var url = auth.AuthURL(state)

func main() {
	log.Println(url)
	var err error = nil
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Println("Copy and paste the url ^")
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", RedirectHandler)
	http.ListenAndServe(":51337", mux)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	token, _ := auth.Token(r.Context(), state, r)
	data, _ := json.Marshal(*token)
	w.Write(data)

}

// {"access_token":"<stufffffffffffffffffffffff>","token_type":"Bearer","refresh_token":"<stufffffffffffffffffffffffff>","expiry":"2023-01-18T01:43:23.2215505-08:00"}
// GOOS=windows GOARCH=amd64 go build cmd/spotifyoauth/main.go
