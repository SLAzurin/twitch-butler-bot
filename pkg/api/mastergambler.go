package api

import (
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/utils"
)

var gamblingState = struct {
	remainingMoney  uint64
	percentageToBet uint64
}{
	remainingMoney:  1,
	percentageToBet: 30,
}

var gambleChan = make(chan struct {
	amt     uint64
	channel string
})

func MasterGambler() {
	if v := os.Getenv("POINTS"); v != "" && os.Getenv("CHANNEL") != "" {
		gamblingState.remainingMoney, _ = strconv.ParseUint(v, 10, 64)
		go func() {
			time.Sleep(5 * time.Second)
			if gamblingState.remainingMoney/100*gamblingState.percentageToBet > 2147483647 {
				*msgChan <- chat("!gamble 2147483647", os.Getenv("CHANNEL"))
			} else {
				*msgChan <- chat("!gamble "+strconv.FormatUint(gamblingState.remainingMoney/100*gamblingState.percentageToBet, 10), os.Getenv("CHANNEL"))
			}
		}()
	}

	go func() {
		for {
			var data = <-gambleChan
			log.Println("GOT DATA TO GAMBLE, WAIT NOW", strconv.FormatUint(data.amt, 10))
			time.Sleep(31 * time.Second)
			if gamblingState.remainingMoney/100*gamblingState.percentageToBet > 2147483647 {
				*msgChan <- chat("!gamble "+strconv.FormatUint(2147483647, 10), data.channel)
			} else {
				*msgChan <- chat("!gamble "+strconv.FormatUint(data.amt, 10), data.channel)
			}
		}
	}()

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
				go processIRCGambler(irc, v, n)
			}
		}
	}
}

func processIRCGambler(irc *IRCConn, incoming string, n int) {
	breakdown := strings.Split(incoming, " ")
	// identity := breakdown[0]
	user := ""
	if len(breakdown) > 1 {
		user = breakdown[1]
	}
	incomingType := ""
	if len(breakdown) > 2 {
		incomingType = breakdown[2]
	}
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
	case incomingType == "PRIVMSG" && user == ":streamlabs!streamlabs@streamlabs.tmi.twitch.tv" && strings.Contains(strings.ToLower(incoming), "@azurindayo") && (strings.HasSuffix(incoming, "Seeds") || strings.HasSuffix(incoming, "Seeds.")):
		logirc.Println("Remaining seeds", brokenMessage[len(brokenMessage)-2])
		gamblingState.remainingMoney, _ = strconv.ParseUint(brokenMessage[len(brokenMessage)-2], 10, 64)
		logirc.Println("GAMBLING MODE ON!", gamblingState.remainingMoney, gamblingState.remainingMoney*100/gamblingState.percentageToBet, strings.Join(brokenMessage, " "))
		gambleChan <- struct {
			amt     uint64
			channel string
		}{amt: gamblingState.remainingMoney / 100 * gamblingState.percentageToBet, channel: incomingChannel}
	}
}
