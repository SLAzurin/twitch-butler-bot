package api

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"golang.org/x/net/websocket"
)

/*
To add a channel, add it to the .env file
*/

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

func ircMain() {
	var err error
	for {
		var msg = make([]byte, 1024)
		var n int
		if n, err = irc.Conn.Read(msg); err != nil {
			log.Println("error when reading websocket msg", err)
			utils.LogToFile("error when reading websocket msg", err, string(debug.Stack()))
			if irc != nil {
				irc.Conn.Close()
				time.Sleep(2 * time.Second)
			}
			*exitCh <- struct{}{}
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

func (irc *IRCConn) IRCMessage(s string) {
	websocket.Message.Send(irc.Conn, s)
}

func chat(s string, channel string) string {
	return "PRIVMSG " + channel + " :" + s
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
	var brokenMessage []string
	if len(breakdown) > 4 {
		brokenMessage = breakdown[4:]
		brokenMessage[0] = brokenMessage[0][1:] // Removes colon from the first character of the full message
	}

	switch {
	case incoming == ":tmi.twitch.tv RECONNECT":
		*exitCh <- struct{}{}
		return
	case strings.HasPrefix(incoming, "PING"):
		*msgChan <- strings.Replace(incoming, "PING", "PONG", 1)
		logirc.Println(strings.Replace(incoming, "PING", "PONG", 1))
	case strings.Contains(incoming, " CLEARCHAT ") || strings.HasPrefix(incoming, "@ban-duration="):
		// handleBan("@ban-duration=1;room-id=254067565;target-user-id=19609203;tmi-sent-ts=1673678936710 :tmi.twitch.tv CLEARCHAT #sangnope :omnoloko")
		handleBan(incoming, incomingChannel)
	case strings.Contains(identity, "custom-reward-id="):
		handleRewards(identity, incomingChannel, user, utils.GetPermissionLevel(utils.IdentityParser(identity)), brokenMessage)
	case len(brokenMessage) > 0 && strings.HasPrefix(brokenMessage[0], "!"):
		handleCommand(incomingChannel, user, utils.GetPermissionLevel(utils.IdentityParser(identity)), brokenMessage)
	case strings.Contains(incoming, "PRIVMSG"):
		handleMessageScan(incomingChannel, user, utils.GetPermissionLevel(utils.IdentityParser(identity)), brokenMessage)
	}
	logirc.Println(incoming)
}

var lastBanTime = map[string]time.Time{}

func handleBan(rawmsg string, channel string) {
	val, err := apidb.RedisDB.Get(context.Background(), channel+"_autounban").Result()
	if err != nil {
		return
	}
	var autoUnbanUsers []string
	json.Unmarshal([]byte(val), &autoUnbanUsers)

	bannedUser := strings.ToLower(rawmsg[strings.LastIndex(rawmsg, ":")+1:])

	toAutoUnban := false
	for _, v := range autoUnbanUsers {
		if strings.ToLower(v) == bannedUser {
			toAutoUnban = true
			break
		}
	}

	if toAutoUnban {
		logirc.Println("Unbanning " + bannedUser)
		time.Sleep(time.Second * 2)
		*msgChan <- chat("/unban "+bannedUser, channel)

		oldTime := time.Now().Add(-6 * time.Second)
		if t, ok := lastBanTime[channel]; ok {
			oldTime = t
		}
		lastBanTime[channel] = time.Now()
		if oldTime.After(time.Now().Add(-5 * time.Second)) {
			return
		}

		// fluff here
		*msgChan <- chat(GetRandomUnbanPhrase(), channel)

	}
}

var randomUnbanPhrase = []string{
	"Y U GOTTA B LIEK DIS LARRY ericareiGun",
	"No unethical bans here ericareiKnife",
	"No ban! ericareiPout",
	"Unbanned. ericareiGiggle",
}

func GetRandomUnbanPhrase() string {
	rand.Seed(time.Now().UnixNano())
	return randomUnbanPhrase[rand.Intn(len(randomUnbanPhrase))]
}
