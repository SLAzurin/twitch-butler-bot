package api

import "strings"

var ModCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#ericarei": {
		"!autosr":   toggleAutoSR,
		"!togglesr": toggleAutoSR,
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
