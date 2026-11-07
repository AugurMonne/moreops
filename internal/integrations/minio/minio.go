package minio

// DevConfig returns default in-cluster MinIO configuration.
type DevConfig struct {
	Port        int
	ConsolePort int
	User        string
	Password    string
}

// DefaultDevConfig returns standard credentials and ports for dev MinIO.
func DefaultDevConfig() DevConfig {
	return DevConfig{
		Port:        9000,
		ConsolePort: 9001,
		User:        "minioadmin",
		Password:    "minioadmin",
	}
}
