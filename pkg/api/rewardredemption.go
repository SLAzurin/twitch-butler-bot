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
var rewardsMap = map[string]func(msgChan *chan string, channel string, actualMessage string){
	"#ericarei=110b2338-fef9-47c1-be96-39363e0b5c87": processSongRequestNightBot,
	"#sangnope=57066ddf-2db9-439f-8a19-561f67c49474": processSongRequestSpotify,
}

func handleRewards(identity string, incomingChannel string, user string, actualMesage string) {
	identityMap := utils.IdentityParser(identity)
	if f, ok := rewardsMap[incomingChannel+"="+(*identityMap)["custom-reward-id"]]; ok {
		f(msgChan, incomingChannel, actualMesage)
		return
	}
	logreward.Println(user+":", (*identityMap)["custom-reward-id"], actualMesage)

}
