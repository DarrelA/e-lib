package entity

type (
	EnvConfig struct {
		AppEnv              string
		Port                string
		PathToSQLSchema     string
		PathToBooksJsonFile string
		PostgresDBConfig    *PostgresDBConfig
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
