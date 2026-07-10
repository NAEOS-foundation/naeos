package config

// Load returns a starter configuration for the sample-specification module.
func Load() Config {
	return Config{Port: 8080}
}
