package configs

type Discord struct {
	Token                  string `env:"DISCORD_AUTHORIZATION_BOT_TOKEN,notEmpty"`
	AuthorizationChannelID string `env:"DISCORD_AUTHORIZATION_CHANNEL_ID,notEmpty"`
}
