package api

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/data"
	"golang.org/x/net/websocket"
)

type IRCConn struct {
	Conn *websocket.Conn
}

var host = "wss://irc-ws.chat.twitch.tv"
var logirc = log.New(os.Stdout, "IRC ", log.Ldate|log.Ltime)
var irc *IRCConn = nil
var connectRetries = 0

var msgChan *chan string

func init() {
	var c = make(chan string)
	msgChan = &c
	go func() {
		for {
			if irc == nil {
				continue
			}
			var s = <-*msgChan
			logirc.Println("Send: " + s)
			irc.IRCMessage(s)
		}
	}()
}

func (irc *IRCConn) IRCMessage(s string) {
	websocket.Message.Send(irc.Conn, s)
}

func chat(s string) string {
	return "PRIVMSG " + data.AppCfg.TwitchChannel + " :" + s
}

func Run(exitCh *chan struct{}) {
	var rawConn *websocket.Conn
	var err error
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

	// Login here
	*msgChan <- "CAP REQ :twitch.tv/membership twitch.tv/tags twitch.tv/commands"
	*msgChan <- "PASS oauth:" + data.AppCfg.TwitchPassword
	*msgChan <- "NICK " + data.AppCfg.TwitchAccount
	*msgChan <- "JOIN " + data.AppCfg.TwitchChannel

	for {
		var msg = make([]byte, 1024)
		var n int
		if n, err = irc.Conn.Read(msg); err != nil {
			break
		}
		stringmsg := string(msg[:n])
		for _, v := range strings.Split(stringmsg, "\r\n") {
			go processIRC(irc, v, n)
		}
	}
}

func processIRC(irc *IRCConn, incoming string, n int) {
	switch {
	case strings.HasPrefix(incoming, "PING"):
		*msgChan <- strings.Replace(incoming, "PING", "PONG", 1)
		logirc.Println(strings.Replace(incoming, "PING", "PONG", 1))
	case strings.HasPrefix(incoming, "@ban-duration="):
		fallthrough
	case strings.Contains(incoming, " CLEARCHAT "):
		// handleBan("@ban-duration=1;room-id=254067565;target-user-id=19609203;tmi-sent-ts=1673678936710 :tmi.twitch.tv CLEARCHAT #sangnope :omnoloko")
		handleBan(incoming)
	default:
		logirc.Println(incoming)
	}
}

func handleBan(msg string) {
	bannedUser := msg[strings.LastIndex(msg, ":")+1:]
	perm := !strings.HasPrefix(msg, "@ban-duration=")

	if strings.Contains(data.AppCfg.AutoUnbans, bannedUser) {
		logirc.Println("Unbanning " + bannedUser)
		if perm {
			*msgChan <- chat("/unban " + bannedUser)
		} else {
			*msgChan <- chat("/untimeout " + bannedUser)
		}
		// fluff here
	}
}
