package api

import (
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"golang.org/x/net/websocket"
)

/*
Main entrypoint
*/
var exitCh *chan struct{}

func Run(_exitCh *chan struct{}) {
	exitCh = _exitCh
	var err error

	// IRC
	var rawConn *websocket.Conn

	for irc == nil {
		time.Sleep(time.Second * time.Duration(connectRetries))
		rawConn, err = websocket.Dial(host, "", "http://localhost/")
		if err != nil {
			if connectRetries > 128 {
				logirc.Println("Last retry took 128s and still didn't reconnect")
				logirc.Println("Force closing")
				*exitCh <- struct{}{}
				return
			}
			logirc.Println("Failed to connect", err)
			logirc.Println("Retrying")
			if connectRetries == 0 {
				connectRetries = 1
			} else {
				connectRetries *= 2
			}
			continue
		}
		connectRetries = 0
		irc = &IRCConn{Conn: rawConn}
	}

	// Login IRC
	*msgChan <- "CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands"
	*msgChan <- "PASS oauth:" + data.AppCfg.TwitchPassword
	*msgChan <- "NICK " + data.AppCfg.TwitchAccount
	*msgChan <- "JOIN " + data.AppCfg.TwitchChannel

	go ircMain()
}
