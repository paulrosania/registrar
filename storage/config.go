package storage

type Config struct {
	Database DatabaseConfig
}

type DatabaseConfig struct {
	User     string
	Password string
	Protocol string
	Host     string
	Port     int
	Database string
	Sslmode  string
	Pool     int
}
