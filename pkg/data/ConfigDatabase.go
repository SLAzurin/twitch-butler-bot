package data

type ConfigDatabase struct {
	TwitchAccountClientID string `env:"TWITCH_ACCOUNT_CLIENT_ID"`
	TwitchChannel         string `env:"TWITCH_CHANNEL"`
	TwitchAccount         string `env:"TWITCH_ACCOUNT"`
	TwitchPassword        string `env:"TWITCH_PASSWORD"`
	TwitchTargetChannel   string `env:"TWITCH_TARGET_CHANNEL"`
	EvilMods              string `env:"EVIL_MODS"`
	AutoUnbans            string `env:"AUTO_UNBANS"`
}

var AppCfg ConfigDatabase
