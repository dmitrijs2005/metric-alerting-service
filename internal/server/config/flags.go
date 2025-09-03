package config

import (
	"flag"
	"os"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
)

func parseFlags(config *Config) {

	// filtering args to leave just values processed by parseFlags
	args := common.FilterArgs(os.Args[1:], []string{"-d", "-a", "-i", "-f", "-k", "-r", "-crypto-key", "-t", "-g"})

	fs := flag.NewFlagSet("main", flag.ContinueOnError)

	fs.StringVar(&config.DatabaseDSN, "d", config.DatabaseDSN, "database dsn")

	fs.StringVar(&config.EndpointAddr, "a", config.EndpointAddr, "address and port to run server")

	var storeInterval int
	fs.IntVar(&storeInterval, "i", int(config.StoreInterval.Seconds()), "saved metric store interval")

	fs.StringVar(&config.FileStoragePath, "f", config.FileStoragePath, "saved metric file storage path")

	fs.StringVar(&config.Key, "k", config.Key, "signing key")
	fs.BoolVar(&config.Restore, "r", config.Restore, "restore saved metrics")

	fs.StringVar(&config.CryptoKey, "crypto-key", config.CryptoKey, "crypto key")

	fs.StringVar(&config.TrustedSubnet, "t", config.TrustedSubnet, "trusted subnet")

	fs.StringVar(&config.GRPCEndpointAddr, "g", config.GRPCEndpointAddr, "GRPC endpoint")

	err := fs.Parse(args)
	if err != nil {
		panic(err)
	}

	config.StoreInterval = time.Duration(storeInterval) * time.Second

}
