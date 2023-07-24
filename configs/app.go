package configs

type App struct {
	Environment string `env:"ENVIRONMENT,notEmpty"`
}

func (c App) IsDevEnvironment() bool {
	return c.Environment == "dev"
}
