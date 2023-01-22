package api

import (
	"strings"
)

var ModCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#ericarei": {
		"!autosr":   toggleAutoSR,
		"!togglesr": toggleAutoSR,
	},
}

var SubsCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#azurindayo": {
		"!sr": commandProcessSongRequestSpotify,
	},
	"#sangnope": {
		"!sr": commandProcessSongRequestSpotify,
	},
}

func handleModCommand(incomingChannel string, user string, acutalMessage string) {
	cmdLen := strings.Index(acutalMessage, " ")
	if cmdLen == -1 {
		cmdLen = len(acutalMessage)
	}
	if m, ok := ModCommands[incomingChannel]; ok {
		if f, ok := m[acutalMessage[:cmdLen]]; ok {
			f(incomingChannel, user, acutalMessage)
			return
		}
	}
	if m, ok := SubsCommands[incomingChannel]; ok {
		if f, ok := m[acutalMessage[:cmdLen]]; ok {
			f(incomingChannel, user, acutalMessage)
			return
		}
	}
}

func handleSubCommand(incomingChannel string, user string, acutalMessage string) {
	cmdLen := strings.Index(acutalMessage, " ")
	if cmdLen == -1 {
		cmdLen = len(acutalMessage)
	}
	if m, ok := SubsCommands[incomingChannel]; ok {
		if f, ok := m[acutalMessage[:cmdLen]]; ok {
			f(incomingChannel, user, acutalMessage)
			return
		}
	}
}

func toggleAutoSR(incomingChannel, user, acutalMessage string) {
	autosr[incomingChannel] = !autosr[incomingChannel]
	if autosr[incomingChannel] {
		*msgChan <- chat("autosr is now on", incomingChannel)
	} else {
		*msgChan <- chat("autosr is now off", incomingChannel)
	}
}

func commandProcessSongRequestSpotify(incomingChannel, user, acutalMessage string) {
	processSongRequestSpotify(msgChan, incomingChannel, acutalMessage)
}
