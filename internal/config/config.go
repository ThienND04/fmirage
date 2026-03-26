package config

import (
	"flag"
	"fmt"
	"reflect"
	"strings"
)

type Config struct {
	TargetURL   string
	Wordlist    string
	Threads     int
	ProxiesFile string
	Timeout     int

	DisableMatcher bool
	MatchCodes     string
	MatchSizes     string
	MatchWords     string
	MatchLines     string

	FilterCodes string
	FilterSizes string
	FilterWords string
	FilterLines string
}

func (c *Config) String() string {
	var builder strings.Builder

	v := reflect.ValueOf(*c)
	typeOfS := v.Type()

	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).IsZero() {
			continue
		}

		fieldName := typeOfS.Field(i).Name
		fieldValue := v.Field(i).Interface()
		builder.WriteString(fmt.Sprintf("%s: %v\n", fieldName, fieldValue))
	}

	return builder.String()
}

func ParseFlags() *Config {
	cfg := &Config{}
	flag.StringVar(&cfg.TargetURL, "u", "", "Target URL to fuzz")
	flag.StringVar(&cfg.Wordlist, "w", "", "Path to the wordlist file")
	flag.IntVar(&cfg.Threads, "t", 10, "Number of concurrent threads")
	flag.StringVar(&cfg.ProxiesFile, "x", "", "Path to the proxies file (optional)")

	// match flags
	flag.BoolVar(&cfg.DisableMatcher, "no-match", false, "Disable match mode (only include results that match criteria)")
	flag.StringVar(&cfg.MatchCodes, "mc", "200,204,301,302,307,401,403", "Status codes to match, separated by commas")
	flag.StringVar(&cfg.MatchSizes, "ms", "", "Response sizes to match, separated by commas")
	flag.StringVar(&cfg.MatchWords, "mw", "", "Response word counts to match, separated by commas")
	flag.StringVar(&cfg.MatchLines, "ml", "", "Response line counts to match, separated by commas")

	// filter flags
	flag.StringVar(&cfg.FilterCodes, "fc", "", "Status codes to filter out, separated by commas")
	flag.StringVar(&cfg.FilterSizes, "fs", "", "Response sizes to filter out, separated by commas")
	flag.StringVar(&cfg.FilterWords, "fw", "", "Word counts to filter out, separated by commas")
	flag.StringVar(&cfg.FilterLines, "fl", "", "Line counts to filter out, separated by commas")

	flag.Parse()
	return cfg
}
