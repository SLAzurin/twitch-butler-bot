package data

type ConfigDatabase struct {
	TwitchChannel          string `env:"TWITCH_CHANNEL"`
	TwitchAccount          string `env:"TWITCH_ACCOUNT"`
	TwitchPassword         string `env:"TWITCH_PASSWORD"`
	AutoUnbans             string `env:"AUTO_UNBANS"`
	AutoUnbanChannels      string `env:"AUTO_UNBAN_CHANNELS"`
	AutoSongRequestChannel string `env:"AUTO_SONG_REQUEST_CHANNEL"`
	AutoSongRequestID      string `env:"AUTO_SONG_REQUEST_ID"`
}

var AppCfg ConfigDatabase
