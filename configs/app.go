package configs

type App struct {
	Environment            string   `env:"ENVIRONMENT,notEmpty"`
	CommunityName          string   `env:"COMMUNITY_NAME" envDefault:"S16"`
	VotingDurationDays     int      `env:"VOTING_DURATION_DAYS" envDefault:"7"`
	RenominationPeriodDays int      `env:"RENOMINATION_PERIOD_DAYS" envDefault:"3"`
	InitialSeeders         []string `env:"INITIAL_SEEDERS" envSeparator:","`
	MembersChatID          int64    `env:"MEMBERS_CHAT_ID"`
	SeedersChatID          int64    `env:"SEEDERS_CHAT_ID"`
}
