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
	PostgresUser           string `env:"POSTGRES_USER"`
	PostgresPassword       string `env:"POSTGRES_PASSWORD"`
	PostgresDB             string `env:"POSTGRES_DB"`
	PostgresPort           string `env:"CLIENT_POSTGRES_PORT"`
	PostgresHost           string `env:"CLIENT_POSTGRES_HOST"`
	RedisPort              string `env:"CLIENT_REDIS_PORT"`
	RedisHost              string `env:"CLIENT_REDIS_HOST"`
	AzuriAIPort            string `env:"AZURIAI_PORT"`
	AzuriAIHost            string `env:"AZURIAI_HOST"`
}

var AppCfg ConfigDatabase
