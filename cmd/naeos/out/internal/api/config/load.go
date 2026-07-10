package config

// Load returns a starter configuration for the api module.
func Load() Config {
	return Config{Port: 8080}
}
