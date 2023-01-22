package utils

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
)

// https://api.twitch.tv/helix/streams?user_login=ericarei

func ChannelIsLive(channel string) (bool, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.twitch.tv/helix/streams?user_login="+channel, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Client-ID", data.AppCfg.TwitchAccountClientID)
	req.Header.Add("Authorization", "Bearer "+data.AppCfg.TwitchPassword)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to subscribe to channel.ban", err)
		return false, err
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	data := data.ApiStreams{}
	json.Unmarshal(respBody, &data)
	if len(data.Data) > 0 {
		return true, nil
	}
	return false, nil
}
