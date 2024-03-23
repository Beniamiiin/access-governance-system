package configs

type App struct {
	Environment        string   `env:"ENVIRONMENT,notEmpty"`
	VotingDurationDays int      `env:"VOTING_DURATION_DAYS" envDefault:"7"`
	InitialSeeders     []string `env:"INITIAL_SEEDERS" envSeparator:","`
	MembersChatID      int64    `env:"MEMBERS_CHAT_ID"`
	SeedersChatID      int64    `env:"SEEDERS_CHAT_ID"`
}
