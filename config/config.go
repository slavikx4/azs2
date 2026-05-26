package config

type Config struct {
	DB DBConfig
}

type DBConfig struct {
	ConnectionString string `envconfig:"DB_CONNECTION_STRING" required:"true"`
}
