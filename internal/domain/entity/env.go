package entity

type (
	EnvConfig struct {
		Port             string
		LogFilePath      string
		PostgresDBConfig *PostgresDBConfig
	}

	PostgresDBConfig struct {
		Username     string
		Password     string
		Host         string
		Port         string
		Name         string
		SslMode      string
		PoolMaxConns string
	}
)
