package config

//var config Config

type Config struct {
	EndpointAddr   string
	ReportInterval int
	PollInterval   int
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
