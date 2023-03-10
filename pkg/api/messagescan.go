package api

import (
	"log"
	"strconv"
	"strings"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
)

var messageScanMap = map[string]func(incomingChannel string, user string, permissionLevel int, brokenMessage []string){
	"#sangnope": func(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
		if strings.Contains(strings.ToLower(strings.Join(brokenMessage, " ")), "dumpy") {
			var dumpyCount int64 = 0
			err := apidb.DB.QueryRow(`UPDATE sangnope set num = (num + 1) where id = 1 returning num`).Scan(&dumpyCount)
			if err != nil {
				*msgChan <- chat("NO DUMPY!?!?!?! ericareiShock2", incomingChannel)
				log.Println(err)
			}
			if dumpyCount%50 == 0 && dumpyCount != 0 {
				*msgChan <- chat(strconv.FormatInt(dumpyCount, 10)+" dumpies ericareiGiggle", incomingChannel)
			}
		}
	},
}

func handleMessageScan(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	if f, ok := messageScanMap[incomingChannel]; ok {
		f(incomingChannel, user, permissionLevel, brokenMessage)
	}
}
