package config

import (
	"flag"
)

type Config struct {
	TargetURL   string
	Wordlist    string
	Threads     int
	ProxiesFile string
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.TargetURL, "u", "", "Target URL to fuzz")
	flag.StringVar(&cfg.Wordlist, "w", "", "Path to the wordlist file")
	flag.IntVar(&cfg.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&cfg.ProxiesFile, "x", "", "Path to the proxies file (optional)")
	flag.Parse()
	return cfg
}
