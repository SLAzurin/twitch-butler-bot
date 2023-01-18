package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"time"

	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

var state = "state-azurindayo-" + strconv.FormatInt(time.Now().Unix(), 10)
var auth = spotifyauth.New(
	spotifyauth.WithRedirectURL("http://localhost:51337"),
	spotifyauth.WithScopes(
		spotifyauth.ScopeUserModifyPlaybackState,
		spotifyauth.ScopeUserReadCurrentlyPlaying,
		spotifyauth.ScopeUserReadPlaybackState,
		spotifyauth.ScopeUserReadRecentlyPlayed,
	),
	// spotifyauth.WithClientID("SUPPLY UR OWN CLIENT ID"),
	// spotifyauth.WithClientSecret("SUPPLY UR OWN CLIENT SECRET"),
)
var url = auth.AuthURL(state)

func main() {

	var err error
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
		log.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", RedirectHandler)
	http.ListenAndServe(":51337", mux)
}

func RedirectHandler(w http.ResponseWriter, r *http.Request) {
	token, _ := auth.Token(r.Context(), state, r)
	data, _ := json.Marshal(*token)
	w.Write(data)
	r.Body.Close()
}

// {"access_token":"<stufffffffffffffffffffffff>","token_type":"Bearer","refresh_token":"<stufffffffffffffffffffffffff>","expiry":"2023-01-18T01:43:23.2215505-08:00"}
// GOOS=windows GOARCH=amd64 go build cmd/spotifyoauth/main.go
