package common

import (
	"flag"
	"os"
)

// jsonConfigEnv retrieves the configuration file path from the CONFIG
// environment variable.
//
// If CONFIG is set and non-empty, its value is returned. Otherwise, an empty
// string is returned.
func JsonConfigEnv() string {
	if envVar, ok := os.LookupEnv("CONFIG"); ok && envVar != "" {
		return envVar
	}
	return ""
}

// jsonConfigFlags inspects command-line arguments and extracts the config file
// path provided via the -c or -config flags.
//
// Only these flags are parsed; other arguments are ignored. This allows the
// application to safely parse its own flags without interfering with flags
// defined by other packages.
//
// If neither -c nor -config is present, an empty string is returned.
func JsonConfigFlags() string {
	var config string

	args := FilterArgs(os.Args[1:], []string{"-c", "-config"})

	fs := flag.NewFlagSet("json", flag.ContinueOnError)
	fs.StringVar(&config, "config", "", "Path to config file")
	fs.StringVar(&config, "c", "", "Path to config file (short)")
	_ = fs.Parse(args)

	return config
}
