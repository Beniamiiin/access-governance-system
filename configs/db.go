package configs

type DB struct {
	URL string `env:"DATABASE_URL,notEmpty"`
}
