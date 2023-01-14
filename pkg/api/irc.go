package api

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/slazurin/twitch-ban-negate/pkg/data"
	"golang.org/x/net/websocket"
)

type IRCConn struct {
	Conn *websocket.Conn
}

var host = "wss://irc-ws.chat.twitch.tv"
var logirc = log.New(os.Stdout, "IRC ", log.Ldate|log.Ltime)
var irc *IRCConn = nil
var connectRetries = 0

var msgChan = make(chan string)

func init() {
	go func() {
		for {
			if irc == nil {
				continue
			}
			irc.IRCMessage(<-msgChan)
		}
	}()
}

func (irc *IRCConn) IRCMessage(s string) {
	websocket.Message.Send(irc.Conn, s)
}

func chat(s string) string {
	return "PRIVMSG " + data.AppCfg.TwitchChannel + ":" + s
}

func runIRC(exitCh *chan struct{}) {
	for irc == nil {
		time.Sleep(time.Second * time.Duration(connectRetries))
		rawConn, err := websocket.Dial(host, "", "http://localhost/")
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

		}
		connectRetries = 0
		irc = &IRCConn{Conn: rawConn}
	}

	// Login here
	msgChan <- "PASS oauth:" + data.AppCfg.TwitchPassword
	msgChan <- "NICK " + data.AppCfg.TwitchAccount
	go func() {
		time.Sleep(3 * time.Second)
		msgChan <- "JOIN " + data.AppCfg.TwitchChannel
		msgChan <- chat("I am connected")
	}()

	var err error
	for {
		var msg = make([]byte, 512)
		var n int
		if n, err = irc.Conn.Read(msg); err != nil {
			break
		}
		go processIRC(irc, msg, n)
	}
}

func processIRC(irc *IRCConn, msg []byte, n int) {
	var incoming = string(msg[:n])

	switch {
	case strings.HasPrefix(incoming, "PING"):
		msgChan <- strings.Replace(incoming, "PING", "PONG", 1)
	default:
		logirc.Println(incoming)
	}
}
