package configs

type AccessGovernanceBot struct {
	Token string `env:"TELEGRAM_ACCESS_GOVERNANCE_BOT_TOKEN,notEmpty"`
}
