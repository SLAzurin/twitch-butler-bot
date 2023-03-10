package api

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"github.com/slazurin/twitch-butler-bot/pkg/apidb"
)

/*
Add commands under each channel here
*/

var commandCoolDowns = map[string]map[string]time.Time{}

var AnyCommands = map[int]func(incomingChannel string, user string, permissionLevel int, brokenMessage []string){
	1: toggleAutoSR,
	2: commandSkipSongSpotify,
	3: commandDumpy,
	6: commandProcessSongRequestSpotify,
	7: commandMapleRanks,
	8: commandDisable,
}

func handleCommand(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	sqlQuery := `select channel_commands.command, special, basic_output, channel_command_perm_overrides.allowed, permission_level, channel_commands.id
	from channel_commands
	  full outer join channel_command_aliases ON channel_command_aliases.channel_command_id = channel_commands.id
	  left join channels on channels.id = channel_commands.channel_id
	  left join channel_command_perm_overrides ON channel_commands.id = channel_command_perm_overrides.channel_command_id and channel_command_perm_overrides.username = $1
	  where (channel_commands.command = $2 or channel_command_aliases.alias = $2) and (channel_commands.channel_id = 0 or channels.channel_name = $3);`
	stmt, err := apidb.DB.Prepare(sqlQuery)
	if err != nil {
		log.Println("Failed 1st prepare at handleCommand", err)
		return
	}
	rows, err := stmt.Query(user, brokenMessage[0], incomingChannel)
	if err != nil {
		log.Println("Failed 1st query at handleCommand", err)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		return
	}
	var rCommand string
	var rSpecial bool
	var rBasicOutputP *string
	var rAllowedP *bool
	var rPermissionLevel int
	var rCommandID int
	err = rows.Scan(&rCommand, &rSpecial, &rBasicOutputP, &rAllowedP, &rPermissionLevel, &rCommandID)
	if err != nil {
		log.Println("Failed 1st Scan at handleCommand", err)
		return
	}
	if rCommand != "!disable" {
		val, err := apidb.RedisDB.Get(context.Background(), incomingChannel+"_disabled_"+rCommand).Result()
		if err != nil && err.Error() != "redis: nil" {
			// *msgChan <- chat("I couldn't check if this command was disabled ericareiThink", incomingChannel)
			log.Println("Redis: Couldn't check if", rCommand, "was disabled", err)
			return
		}
		if val == "true" {
			return
		}
	}

	if rAllowedP != nil {
		if !*rAllowedP {
			// explicitly not allowed
			return
		}
		// explicitly allowed
	} else {
		if permissionLevel < rPermissionLevel {
			// not enough perms
			return
		}
	}
	if !rSpecial {
		*msgChan <- chat(*rBasicOutputP, incomingChannel)
		return
	}

	if f, ok := AnyCommands[rCommandID]; ok {
		f(incomingChannel, user, permissionLevel, brokenMessage)
	}
}

func toggleAutoSR(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	val, err := apidb.RedisDB.Get(context.Background(), incomingChannel+"_!autosr").Result()
	if err != nil && err.Error() != "redis: nil" {
		log.Println("redis error get", brokenMessage[0], err.Error())
		return
	}
	if val == "" {
		val = "true"
	}
	var b bool
	json.Unmarshal([]byte(val), &b)
	b = !b
	newVal, _ := json.Marshal(b)
	apidb.RedisDB.Set(context.Background(), incomingChannel+"_"+brokenMessage[0], string(newVal), 0)
	if b {
		*msgChan <- chat("autosr is now on", incomingChannel)
	} else {
		*msgChan <- chat("autosr is now off", incomingChannel)
	}
}

func commandProcessSongRequestSpotify(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	processSongRequestSpotify(msgChan, incomingChannel, permissionLevel, brokenMessage)
}

func commandDumpy(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	if _, ok := commandCoolDowns[incomingChannel]; !ok {
		commandCoolDowns[incomingChannel] = map[string]time.Time{
			"!dumpy": time.Now().Add(-1 * time.Second),
		}
	}
	if commandCoolDowns[incomingChannel]["!dumpy"].Add(10 * time.Second).After(time.Now()) {
		return
	}
	commandCoolDowns[incomingChannel]["!dumpy"] = time.Now()
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
func commandDisable(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
}

func commandMapleRanks(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	if len(brokenMessage) > 1 {
		*msgChan <- chat("https://mapleranks.com/u/"+brokenMessage[1], incomingChannel)
	}
}
