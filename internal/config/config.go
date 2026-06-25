package config

type Config struct {
	Listen  string
	DataDir string
}

func Load() *Config {
	return &Config{
		Listen:  ":8080",
		DataDir: "./data",
	}
}

