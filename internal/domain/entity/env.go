package entity

type (
	EnvConfig struct {
		AppEnv              string
		Port                string
		PathToSQLSchema     string
		PathToBooksJsonFile string
		PostgresDBConfig    *PostgresDBConfig
		OAuth2Config        *OAuth2Config
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

	OAuth2Config struct {
		GoogleRedirectURL  string
		GoogleClientID     string
		GoogleClientSecret string
		Scopes             []string
	}
)
