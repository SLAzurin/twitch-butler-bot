package api

// import (
// 	"bytes"
// 	"encoding/json"
// 	"io"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strings"

// 	"github.com/slazurin/twitch-butler-bot/pkg/data"
// 	"golang.org/x/net/websocket"
// )

// var location = "wss://eventsub-beta.wss.twitch.tv/ws"

// // var newSessionID = ""
// var sessionID = ""
// var msgHistory = make(map[string]data.EventSubMessage)
// var logeventsub = log.New(os.Stdout, "EVENTSUB ", log.Ldate|log.Ltime)

// func Run(exitCh *chan struct{}) {
// 	defer func() { *exitCh <- struct{}{} }()

// 	go runIRC(exitCh)

// 	logeventsub.Println("Connecting to", location)

// 	ws, err := websocket.Dial(location, "", "http://localhost/")
// 	if err != nil {
// 		logeventsub.Println("Failed to connect to wss")
// 		return
// 	}
// 	for {
// 		var msg = make([]byte, 1024)
// 		var n int
// 		if n, err = ws.Read(msg); err != nil {
// 			break
// 		}
// 		go processEventSub(ws, msg, n)
// 	}
// }

// func processEventSub(ws *websocket.Conn, msg []byte, n int) {
// 	defer func() { go clearHistory() }()
// 	var esMsg = data.EventSubMessage{}
// 	err := json.Unmarshal(msg[:n], &esMsg)
// 	if err != nil {
// 		logeventsub.Println("Failed to unmarshal", string(msg[:n]), err)
// 		return
// 	}
// 	if _, ok := msgHistory[esMsg.Metadata.MessageID]; ok {
// 		return
// 	}
// 	msgHistory[esMsg.Metadata.MessageID] = esMsg

// 	switch {
// 	case esMsg.Metadata.MessageType == "session_welcome":
// 		handleSessionWelcome(ws, esMsg)
// 	case esMsg.Metadata.MessageType == "notification" && esMsg.Metadata.SubscriptionType == "channel.ban":
// 		handleBan(ws, esMsg)
// 	case esMsg.Metadata.MessageType == "session_keepalive":
// 		// should not do anything
// 	default:
// 		logeventsub.Println("not processed", string(msg[:n]), err)
// 	}
// }

// func clearHistory() {
// 	// TODO: for loop through history and wipe all history >10m
// }

// func handleSessionWelcome(ws *websocket.Conn, esm data.EventSubMessage) {
// 	if esm.Payload.Session.Status != "connected" {
// 		logeventsub.Println("Failed to connect at session_welcome")
// 		return
// 	}

// 	if sessionID == "" {
// 		sessionID = esm.Payload.Session.ID
// 	}

// 	// This never errors below.
// 	body, _ := json.Marshal(data.EventSubRequestWebsocket{
// 		Type:    "channel.ban",
// 		Version: "1",
// 		Condition: struct {
// 			BroadcasterUserID string "json:\"broadcaster_user_id\""
// 		}{
// 			BroadcasterUserID: data.AppCfg.TwitchTargetChannel,
// 		},
// 		Transport: struct {
// 			Method    string "json:\"method\""
// 			SessionID string "json:\"session_id\""
// 		}{
// 			Method:    "websocket",
// 			SessionID: sessionID,
// 		},
// 	})

// 	client := &http.Client{}

// 	logeventsub.Println(string(body))

// 	req, err := http.NewRequest("POST", "https://api.twitch.tv/helix/eventsub/subscriptions", bytes.NewBuffer(body))
// 	if err != nil {
// 		logeventsub.Println("Failed to create req client to sub topics.")
// 		return
// 	}
// 	req.Header.Add("Client-ID", data.AppCfg.TwitchAccountClientID)
// 	req.Header.Add("Authorization", "Bearer "+data.AppCfg.TwitchPassword)
// 	req.Header.Add("Content-Type", "application/json")

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		logeventsub.Println("Failed to subscribe to channel.ban", err)
// 		return
// 	}
// 	respBody, _ := io.ReadAll(resp.Body)
// 	resp.Body.Close()
// 	logeventsub.Println("Subscribe to channel.ban", string(respBody))

// }

// func handleBan(ws *websocket.Conn, esm data.EventSubMessage) {
// 	if strings.Contains(data.AppCfg.EvilMods, esm.Payload.Event.ModeratorUserLogin) && strings.Contains(data.AppCfg.AutoUnbans, esm.Payload.Event.UserLogin) {
// 		logeventsub.Println("Unbanning " + esm.Payload.Event.UserLogin + " banned by " + esm.Payload.Event.ModeratorUserLogin)
// 		if esm.Payload.Event.IsPermanant {
// 			*msgChan <- chat("/unban " + esm.Payload.Event.UserLogin)
// 		} else {
// 			*msgChan <- chat("/untimeout " + esm.Payload.Event.UserLogin)
// 		}
// 		// fluff here
// 	}
// }
