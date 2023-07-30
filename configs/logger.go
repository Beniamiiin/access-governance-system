package configs

type Logger struct {
	AppName string
	URL     string `env:"LOKI_URL"`
}
