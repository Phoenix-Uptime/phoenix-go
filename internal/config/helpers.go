package config

func GetServerConfig() (string, string) {
	return config.Server.Host, config.Server.Port
}

func GetDatabaseDriver() string {
	return config.Database.Driver
}

func GetSQLitePath() string {
	return config.Database.SQLite.Path
}

func GetPostgresConfig() (string, string, string, string, string) {
	if config.Database.Driver != "postgres" {
		return "", "", "", "", ""
	}
	pg := config.Database.Postgres
	return pg.Host, pg.Port, pg.User, pg.Password, pg.DBName
}
