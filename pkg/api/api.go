package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/slazurin/twitch-ban-negate/pkg/data"
	"golang.org/x/net/websocket"
)

var messageCh = make(chan string)
var location = "wss://eventsub-beta.wss.twitch.tv/ws"
var newSessionID = ""
var sessionID = ""
var msgHistory = make(map[string]data.EventSubMessage)

func Run(exitCh *chan struct{}) {
	defer func() { *exitCh <- struct{}{} }()
	log.Println("Connecting to", location)

	ws, err := websocket.Dial(location, "", "http://localhost/")
	if err != nil {
		log.Println("Failed to connect to wss")
		return
	}
	for {
		var msg = make([]byte, 512)
		var n int
		if n, err = ws.Read(msg); err != nil {
			break
		}
		go processEventSub(ws, msg, n)
	}
}

func processEventSub(ws *websocket.Conn, msg []byte, n int) {
	defer func() { go clearHistory() }()
	var esMsg = data.EventSubMessage{}
	err := json.Unmarshal(msg[:n], &esMsg)
	if err != nil {
		log.Println("Failed to unmarshal", string(msg[:]), err)
		return
	}
	// TODO: Add if statement to check if msgID already in there before processing
	msgHistory[esMsg.Metadata.MessageID] = esMsg

	switch esMsg.Metadata.MessageType {
	case "session_welcome":
		handleSessionWelcome(ws, esMsg)
	default:
		esMsgStr, err := json.Marshal(esMsg)
		log.Println("not processed", string(esMsgStr), err)
	}
}

func clearHistory() {
	// TODO: for loop through history and wipe all history >10m
}

func handleSessionWelcome(ws *websocket.Conn, esm data.EventSubMessage) {
	if esm.Payload.Session.Status != "connected" {
		log.Println("Failed to connect at session_welcome")
		return
	}

	if sessionID == "" {
		sessionID = esm.Payload.Session.ID
	}

	body, err := json.Marshal(data.EventSubRequestWebsocket{
		Type:    "channel.ban",
		Version: "1",
		Condition: struct {
			BroadcasterUserID string "json:\"broadcaster_user_id\""
		}{
			BroadcasterUserID: data.AppCfg.TwitchTargetChannel,
		}, 
		Transport: struct {
			Method    string "json:\"method\""
			SessionID string "json:\"session_id\""
		}{
			Method:    "websocket",
			SessionID: sessionID,
		},
	})

	client := &http.Client{}

	log.Println(string(body))

	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(body))
	if err != nil {
		log.Println("Failed to create req client to sub topics.")
		return
	}
	req.Header.Add("Client-ID", data.AppCfg.TwitchAccountClientID)
	req.Header.Add("Authorization", "Bearer "+data.AppCfg.TwitchPassword)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to subscribe to channel.ban", err)
		return
	}
	respBody, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	log.Println("Subscribe to channel.ban", string(respBody))

}
