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
var autosr = true

var msgChan *chan string
var rewardsMap *map[string]string

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

func chat(s string, channel string) string {
	return "PRIVMSG " + channel + " :" + s
}

func Run(exitCh *chan struct{}) {
	var rawConn *websocket.Conn
	var err error
	rewardsMap = IdentityParser(data.AppCfg.AutoSongRequestID)

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
			if v != "" {
				go processIRC(irc, v, n)
			}
		}
	}
}

func processIRC(irc *IRCConn, incoming string, n int) {
	breakdown := strings.Split(incoming, " ")
	identity := breakdown[0]
	user := ""
	if len(breakdown) > 1 {
		user = breakdown[1]
	}
	// incomingType := ""
	// if len(breakdown) > 1 {
	// 	incomingType = breakdown[2]
	// }
	incomingChannel := ""
	if len(breakdown) > 3 {
		incomingChannel = breakdown[3]
	}
	actualMessage := ""
	if len(breakdown) > 4 {
		actualMessage = strings.Join(breakdown[4:], " ")[1:]
	}

	switch {
	case strings.HasPrefix(incoming, "PING"):
		*msgChan <- strings.Replace(incoming, "PING", "PONG", 1)
		logirc.Println(strings.Replace(incoming, "PING", "PONG", 1))
	case strings.Contains(incoming, " CLEARCHAT ") || strings.HasPrefix(incoming, "@ban-duration="):
		// handleBan("@ban-duration=1;room-id=254067565;target-user-id=19609203;tmi-sent-ts=1673678936710 :tmi.twitch.tv CLEARCHAT #sangnope :omnoloko")
		handleBan(incoming, incomingChannel)
	case strings.Contains(identity, "custom-reward-id="):
		handleRewards(identity, incomingChannel, user, actualMessage)
		// bts blood sweat tears
	case strings.Contains(identity, "mod=1") && strings.HasPrefix(actualMessage, "!"):
		handleCommand(incomingChannel, user, actualMessage)
	default:
		logirc.Println(incoming)
	}
}

func IdentityParser(identity string) *map[string]string {
	r := make(map[string]string)
	if !strings.Contains(identity, ";") {
		v := strings.Split(identity, "=")
		r[v[0]] = v[1]
		return &r
	}

	for _, group := range strings.Split(identity, ";") {
		v := strings.Split(group, "=")
		r[v[0]] = v[1]
	}
	return &r
}

func handleRewards(identity string, incomingChannel string, user string, actualMesage string) {
	identityMap := IdentityParser(identity)
	if (*identityMap)["custom-reward-id"] != (*rewardsMap)[incomingChannel] {
		return
	}
	// Per channel implementation. For now only Erica's
	if autosr {
		*msgChan <- chat("!sr "+actualMesage, incomingChannel)
	}
}

func handleCommand(incomingChannel string, user string, acutalMessage string) {
	if strings.HasPrefix(acutalMessage, "!autosr") || strings.HasPrefix(acutalMessage, "!togglesr") {
		autosr = !autosr
		if autosr {
			*msgChan <- chat("autosr is now on", incomingChannel)
		} else {
			*msgChan <- chat("autosr is now off", incomingChannel)
		}
	}
}

func handleBan(rawmsg string, channel string) {
	if !strings.Contains(data.AppCfg.AutoUnbanChannels, channel) {
		return
	}
	bannedUser := rawmsg[strings.LastIndex(rawmsg, ":")+1:]
	perm := !strings.HasPrefix(rawmsg, "@ban-duration=")

	if strings.Contains(data.AppCfg.AutoUnbans, bannedUser) {
		logirc.Println("Unbanning " + bannedUser)
		if perm {
			*msgChan <- chat("/unban "+bannedUser, channel)
		} else {
			*msgChan <- chat("/untimeout "+bannedUser, channel)
		}
		// fluff here
	}
}
