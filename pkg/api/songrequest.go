package api

var autosr = map[string]bool{
	"#ericarei": true,
}

func processSongRequestNightBot(msgChan *chan string, channel string, actualMessage string) {
	if autosr[channel] {
		*msgChan <- chat("!sr "+actualMessage, channel)
	}
}

func processSongRequestSpotify(msgChan *chan string, channel string, actualMessage string) {
	if !autosr[channel] {
		return
	}
}
