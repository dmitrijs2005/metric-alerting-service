package config

//var config Config

type Config struct {
	EndpointAddr    string
	StoreInterval   int
	FileStoragePath string
	Restore         bool
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
