package configs

type Bot struct {
	Token         string `env:"TELEGRAM_ACCESS_GOVERNANCE_BOT_TOKEN,notEmpty"`
	UpdateTimeout int    `env:"TELEGRAM_BOT_UPDATE_TIMEOUT" envDefault:"60"`
}
