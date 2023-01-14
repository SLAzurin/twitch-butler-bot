package data

type ConfigDatabase struct {
	TwitchChannel         string `env:"TWITCH_CHANNEL"`
	TwitchAccount         string `env:"TWITCH_ACCOUNT"`
	TwitchPassword        string `env:"TWITCH_PASSWORD"`
	AutoUnbans            string `env:"AUTO_UNBANS"`
}

var AppCfg ConfigDatabase
