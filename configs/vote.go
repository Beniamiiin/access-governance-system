package configs

type VoteAPI struct {
	URL string `env:"VOTE_API_URL" envDefault:"http://vot-bot-api:8000"`
}
