package api

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
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
		"!skip":     commandSkipSongSpotify,
		"!next":     commandSkipSongSpotify,
	},
}

var SubsCommands = map[string]map[string]func(incomingChannel string, user string, acutalMessage string){
	"#azurindayo": {
		"!sr": commandProcessSongRequestSpotify,
	},
	"#sangnope": {
		"!sr":    commandProcessSongRequestSpotify,
		"!dumpy": commandDumpy,
	},
}

func setAnyChannelCommands() {
	anyChannelCommands = map[string]func(incomingChannel string, user string, isMod bool, acutalMessage string){
		"!mr":       commandMapleRanks,
		"!disable":  commandDisable,
		"!help":     commandHelp,
		"!commands": commandHelp,
		"!azuribot": commandAzuribot,
		"!azuriai":  commandAzuriAI,
	}
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
		if !strings.HasPrefix(acutalMessage, "!azuriai ") {
			return
		}
		log.Println(acutalMessage[:cmdLen])
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

func commandAzuribot(incomingChannel string, user string, isMod bool, acutalMessage string) {
	if isMod {
		*msgChan <- chat("desuwa ericareiLurk", incomingChannel)
	}
}

func commandAzuriAI(incomingChannel string, user string, isMod bool, acutalMessage string) {
	query := acutalMessage[9:]
	payload := struct {
		Content string `json:"content"`
	}{
		Content: query,
	}
	jsonPayload, _ := json.Marshal(payload)
	log.Println("Sending", string(jsonPayload))
	req, err := http.NewRequest("POST", "http://localhost:3000/azuriai", bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}

	var response struct {
		Result string `json:"result"`
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Println(string(body))
		*msgChan <- chat("I-I don't know what you're talking about! ericareiPout", incomingChannel)
		return
	}

	json.Unmarshal(body, &response)

	*msgChan <- chat(response.Result, incomingChannel)

	resp.Body.Close()
}

func commandDumpy(incomingChannel string, user string, acutalMessage string) {
	if commandCoolDowns["!dumpy"].Add(10 * time.Second).After(time.Now()) {
		return
	}
	commandCoolDowns["!dumpy"] = time.Now()
	var dumpyCount int64 = 0
	err := apidb.DB.QueryRow(`SELECT num FROM sangnope WHERE id = 1`).Scan(&dumpyCount)
	if err != nil {
		*msgChan <- chat("NO DUMPY!?!?!?! ericareiShock2", incomingChannel)
		log.Println(err)
		return
	}
	*msgChan <- chat(strconv.FormatInt(dumpyCount, 10)+" dumpies ericareiGiggle", incomingChannel)
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

// TODO: commandCooldowns isnt channel restricted, it is global. it can work in both Sang's and Erica's ch...
var commandCoolDowns = map[string]time.Time{
	"!help":  time.Now().Add(-10 * time.Second),
	"!dumpy": time.Now().Add(-10 * time.Second),
}

func commandHelp(channel string, user string, isMod bool, actualMessage string) {
	if commandCoolDowns["!help"].Add(10 * time.Second).After(time.Now()) {
		return
	}
	commandCoolDowns["!help"] = time.Now()
	*msgChan <- chat("https://gist.github.com/SLAzurin/f77a54a22bdd0a70ec2d81938d432944", channel)
}
