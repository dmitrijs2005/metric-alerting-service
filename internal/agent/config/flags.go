package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
)

func parseFlags(config *Config) {

	// filtering args to leave just values processed by parseFlags
	args := common.FilterArgs(os.Args[1:], []string{"-a", "-r", "-p", "-k", "-l", "-crypto-key", "-g"})

	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&config.EndpointAddr, "a", config.EndpointAddr, "address and port to run server")
	reportInterval := fs.Int("r", int(config.ReportInterval.Seconds()), "report interval (in seconds)")
	pollInterval := fs.Int("p", int(config.PollInterval.Seconds()), "poll interval (in seconds)")
	fs.StringVar(&config.Key, "k", config.Key, "signing key")
	sendRateLimit := fs.Int("l", config.SendRateLimit, "sending rate limit")
	fs.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "crypto key")

	fs.BoolVar(&config.UseGRPC, "g", config.UseGRPC, "use grpc")

	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	config.ReportInterval = time.Duration(*reportInterval) * time.Second
	config.PollInterval = time.Duration(*pollInterval) * time.Second

	config.SendRateLimit = *sendRateLimit

}
