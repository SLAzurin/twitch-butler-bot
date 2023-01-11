package api

import (
	"log"

	"golang.org/x/net/websocket"
)

var messageCh = make(chan string)

var location = "wss://eventsub-beta.wss.twitch.tv/ws"

func Run(exitCh *chan struct{}) {
	log.Println("Connecting to", location)

	ws, err := websocket.Dial(location, "", "http://localhost/")
	for {
		var msg = make([]byte, 512)
		var n int
		if n, err = ws.Read(msg); err != nil {
			break
		}
		log.Printf("Received: %s.\n", msg[:n])
		log.Println(len(msg[:n]))
	}
	*exitCh <- struct{}{}
}
