package config

//var config Config

type Config struct {
	EndpointAddr string
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
