package api

import (
	"bytes"
	"context"
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

var commandCoolDowns = map[string]map[string]time.Time{}

var AnyCommands = map[int]func(incomingChannel string, user string, permissionLevel int, brokenMessage []string){
	1:  toggleAutoSR,
	2:  commandSkipSongSpotify,
	3:  commandDumpy,
	6:  commandProcessSongRequestSpotify,
	7:  commandMapleRanks,
	8:  commandDisable,
	11: commandAzuriAI,
}

func handleCommand(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	sqlQuery := `select channel_commands.command, special, basic_output, channel_command_perm_overrides.allowed, permission_level, channel_commands.id, channel_commands.cooldown
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
	var rCooldown int
	err = rows.Scan(&rCommand, &rSpecial, &rBasicOutputP, &rAllowedP, &rPermissionLevel, &rCommandID, &rCooldown)
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
	if _, ok := commandCoolDowns[incomingChannel]; !ok {
		commandCoolDowns[incomingChannel] = map[string]time.Time{}
	}
	if v, ok := commandCoolDowns[incomingChannel][rCommand]; ok {
		if time.Now().Add(-1 * time.Second * time.Duration(rCooldown)).Before(v) {
			return
		}
	}
	commandCoolDowns[incomingChannel][rCommand] = time.Now()

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

func commandAzuriAI(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	query := brokenMessage[1:]
	payload := struct {
		Content string `json:"content"`
	}{
		Content: strings.Join(query, " "),
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

func commandDumpy(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	var dumpyCount int64 = 0
	err := apidb.DB.QueryRow(`select data::text::bigint from channel_data WHERE channel_id = 2 and id = '!dumpy'`).Scan(&dumpyCount)
	if err != nil {
		*msgChan <- chat("NO DUMPY!?!?!?! ericareiShock2", incomingChannel)
		log.Println(err)
		return
	}
	*msgChan <- chat(strconv.FormatInt(dumpyCount, 10)+" dumpies ericareiGiggle", incomingChannel)
}

// Disable will disable a command from `anyChannelCommands`
func commandDisable(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	// find non alias cmd from brokenMessage[1]
	var rCommand string
	err := apidb.DB.QueryRow(`select command from channel_commands full outer join channel_command_aliases on channel_commands.id = channel_command_aliases.channel_command_id where command = $1 or alias = $1`, brokenMessage[1]).Scan(&rCommand)
	if err != nil {
		*msgChan <- chat("Couldn't find "+brokenMessage[1]+" ericareiCry", incomingChannel)
		return
	}
	// set in redis
	val, err := apidb.RedisDB.Get(context.Background(), incomingChannel+"_disabled_"+rCommand).Result()
	if err != nil && err.Error() != "redis: nil" {
		*msgChan <- chat("500: I couldn't check if this command was disabled ericareiThink", incomingChannel)
		return
	}
	if val == "" {
		val = "false"
	}
	var b bool
	json.Unmarshal([]byte(val), &b)
	b = !b
	newVal, _ := json.Marshal(b)
	apidb.RedisDB.Set(context.Background(), incomingChannel+"_disabled_"+rCommand, string(newVal), 0)
	if b {
		*msgChan <- chat(rCommand+" is now disabled", incomingChannel)
	} else {
		*msgChan <- chat(rCommand+" is now enabled", incomingChannel)
	}
}

func commandMapleRanks(incomingChannel string, user string, permissionLevel int, brokenMessage []string) {
	if len(brokenMessage) > 1 {
		*msgChan <- chat("https://mapleranks.com/u/"+brokenMessage[1], incomingChannel)
	}
}
