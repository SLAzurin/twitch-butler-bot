package data

type ConfigDatabase struct {
	TwitchChannel          string `env:"TWITCH_CHANNEL"`
	TwitchAccountClientID  string `env:"TWITCH_ACCOUNT_CLIENT_ID"`
	TwitchAccount          string `env:"TWITCH_ACCOUNT"`
	TwitchPassword         string `env:"TWITCH_PASSWORD"`
	AutoUnbans             string `env:"AUTO_UNBANS"`
	AutoUnbanChannels      string `env:"AUTO_UNBAN_CHANNELS"`
	AutoSongRequestChannel string `env:"AUTO_SONG_REQUEST_CHANNEL"`
	SpotifyID              string `env:"SPOTIFY_ID"`
	SpotifySecret          string `env:"SPOTIFY_SECRET"`
}

var AppCfg ConfigDatabase
