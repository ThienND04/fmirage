package fuzzer

import (
	"fmirage/internal/config"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type Fuzzer struct {
	Cfg      *config.Config
	Wordlist []string
	Client   *http.Client
}

func NewFuzzer(cfg *config.Config, wordlist []string) *Fuzzer {
	client := NewClient(cfg)
	return &Fuzzer{
		Cfg:      cfg,
		Wordlist: wordlist,
		Client:   client,
	}
}

func (f *Fuzzer) Start() {
	fmt.Printf("Starting fuzzing with %d threads...\n", f.Cfg.Threads)
	jobs := make(chan string, len(f.Wordlist))
	results := make(chan Result, len(f.Wordlist))
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < f.Cfg.Threads; i++ {
		wg.Add(1)
		go f.worker(i, jobs, results, &wg)
	}

	go func() {
		for _, word := range f.Wordlist {
			var targetURL string
			if strings.Contains(f.Cfg.TargetURL, "FUZZ") {
				targetURL = strings.ReplaceAll(f.Cfg.TargetURL, "FUZZ", word)
			} else {
				targetURL = strings.TrimRight(f.Cfg.TargetURL, "/") + "/" + word
			}
			jobs <- targetURL
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("---------------------------------------------------")
	foundCount := 0
	for res := range results {
		fmt.Printf("%s\n", res.String())
		foundCount++
	}
	fmt.Println("---------------------------------------------------")
	fmt.Printf("[*] Quét hoàn tất! Tìm thấy %d kết quả.\n", foundCount)
}
