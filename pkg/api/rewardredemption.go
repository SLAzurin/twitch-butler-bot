package api

import (
	"log"
	"os"
	"strings"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
	"github.com/slazurin/twitch-butler-bot/pkg/utils"
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
	// logreward.Println(user+":", (*identityMap)["custom-reward-id"], brokenMessage)
	rewardID := (*identityMap)["custom-reward-id"]
	var rewardName string
	err := apidb.DB.QueryRow(`select reward_name from channel_rewards left join channels on channels.id = channel_rewards.channel_id where channels.channel_name = $1 and reward_id = $2`, incomingChannel, rewardID).Scan(&rewardName)
	if err != nil && !strings.Contains(err.Error(), "ErrNoRows") {
		logreward.Println("Error when fetching db at handleRewards")
		return
	}
	if err != nil {
		// No rows
		return
	}
	
	

	if f, ok := rewardsMap[rewardName]; ok {
		f(msgChan, incomingChannel, permissionLevel, brokenMessage)
	}
}
