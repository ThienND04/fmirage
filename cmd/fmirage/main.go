package main

import (
	"fmirage/internal/config"
	"fmirage/internal/utils"

	"fmt"
)

func main() {
	fmt.Println("   Welcome to Fmirage")
	cfg := config.ParseFlags()
	fmt.Printf("Target URL: %s\n", cfg.TargetURL)
	fmt.Printf("Wordlist: %s\n", cfg.Wordlist)
	fmt.Printf("Threads: %d\n", cfg.Threads)
	if cfg.ProxiesFile != "" {
		fmt.Printf("Proxies File: %s\n", cfg.ProxiesFile)
	} else {
		fmt.Println("No proxies file provided.")
	}

	words, err := utils.ReadLines(cfg.Wordlist)
	if err != nil {
		fmt.Printf("Error reading wordlist: %v\n", err)
		return
	}
	fmt.Printf("Loaded %d words from the wordlist.\n", len(words))
}
