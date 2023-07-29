package configs

type App struct {
	Environment              string `env:"ENVIRONMENT,notEmpty"`
	CommunityName            string `env:"COMMUNITY_NAME,notEmpty"`
	VotingDurationDays       int    `env:"VOTING_DURATION_DAYS,notEmpty"`
	RenominationPeriodMonths int    `env:"RENOMINATION_PERIOD_MONTHS,notEmpty"`
}

func (c App) IsDevEnvironment() bool {
	return c.Environment == "dev"
}
