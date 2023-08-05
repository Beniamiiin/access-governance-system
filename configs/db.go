package configs

type DB struct {
	URL string `env:"DB_URL,notEmpty"`
}
