package api

import (
	"strings"
)

/*
Add commands under each channel here
*/

var ModCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#ericarei": {
		"!autosr":   toggleAutoSR,
		"!togglesr": toggleAutoSR,
	},
	"#sangnope": {
		"!autosr":   toggleAutoSR,
		"!togglesr": toggleAutoSR,
	},
}

var SubsCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#azurindayo": {
		"!sr": commandProcessSongRequestSpotify,
	},
	"#sangnope": {
		"!sr":   commandProcessSongRequestSpotify,
		"!skip": commandSkipSongSpotify,
		"!next": commandSkipSongSpotify,
	},
}

var disabledAnyCommands = map[string]map[string]struct{}{}

var anyChannelCommands map[string]func(incomingChannel string, user string, isMod bool, acutalMessage string)

func handleAnyCommand(incomingChannel string, user string, isMod bool, acutalMessage string) {
	cmdLen := strings.Index(acutalMessage, " ")
	if cmdLen == -1 {
		cmdLen = len(acutalMessage)
	}
	if _, ok := disabledAnyCommands[incomingChannel][acutalMessage[:cmdLen]]; ok {
		return
	}
	if f, ok := anyChannelCommands[acutalMessage[:cmdLen]]; ok {
		f(incomingChannel, user, isMod, acutalMessage)
		return
	}
}

func handleModCommand(incomingChannel string, user string, isMod bool, acutalMessage string) {
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
	handleSubCommand(incomingChannel, user, isMod, acutalMessage)
}

func handleSubCommand(incomingChannel string, user string, isMod bool, acutalMessage string) {
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
	handleAnyCommand(incomingChannel, user, isMod, acutalMessage)
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

// Disable will disable a command from `anyChannelCommands`
func commandDisable(incomingChannel string, user string, isMod bool, acutalMessage string) {
	if !isMod {
		return
	}
	argv := strings.Split(acutalMessage, " ")

	if len(argv) <= 1 {
		*msgChan <- chat("No command to disable... ericareiCry", incomingChannel)
		return
	}

	cmd := argv[1]
	if !strings.HasPrefix(cmd, "!") {
		cmd = "!" + cmd
	}

	if _, ok := anyChannelCommands[cmd]; !ok {
		*msgChan <- chat("Didn't find that command... ericareiCry", incomingChannel)
		return
	}

	if _, ok := disabledAnyCommands[incomingChannel]; !ok {
		disabledAnyCommands[incomingChannel] = map[string]struct{}{}
	}
	if _, ok := disabledAnyCommands[incomingChannel][cmd]; ok {
		delete(disabledAnyCommands[incomingChannel], cmd)
		*msgChan <- chat("Enabled "+cmd+" ericareiHeart", incomingChannel)
	} else {
		disabledAnyCommands[incomingChannel][cmd] = struct{}{}
		*msgChan <- chat("Disabled "+cmd+" ericareiKnife", incomingChannel)
	}
}

func commandMapleRanks(incomingChannel string, user string, isMod bool, acutalMessage string) {
	argv := strings.Split(acutalMessage, " ")

	if len(argv) > 1 {
		*msgChan <- chat("https://mapleranks.com/u/"+argv[1], incomingChannel)
	}

}
