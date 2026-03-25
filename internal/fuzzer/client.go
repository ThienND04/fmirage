package fuzzer

import (
	"fmirage/internal/config"
	"net/http"
	"time"
)

func NewClient(cfg *config.Config) *http.Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		DisableKeepAlives:   true, // Disable keep-alives to prevent connection reuse
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(cfg.Timeout) * time.Millisecond,
	}
}
