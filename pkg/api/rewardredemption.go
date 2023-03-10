package api

import (
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
	"log"
	"os"
)

/*
Add reward redemptions:
1) note down the uuid for the reward id
2) add it to rewardsMap and add the func to hook on it.
*/

var logreward = log.New(os.Stdout, "REWARD ", log.Ldate|log.Ltime)
var rewardsMap = map[string]func(msgChan *chan string, channel string, permissionLevel int, brokenMessage []string){
	"sr_nightbot": processSongRequestNightBot,
	"sr_spotify":  processSongRequestSpotify,
}

func handleRewards(identity string, incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	identityMap := utils.IdentityParser(identity)
	logreward.Println(user+":", (*identityMap)["custom-reward-id"], brokenMessage)

	if f, ok := rewardsMap[(*identityMap)["custom-reward-id"]]; ok {
		f(msgChan, incomingChannel, permissionLevel, brokenMessage)
	}
}
