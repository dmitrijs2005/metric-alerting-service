package config

var config Config

type Config struct {
	EndpointAddr string
}

func LoadConfig() *Config {
	parseFlags()
	parseEnv()
	return &config
}
