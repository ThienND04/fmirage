package fuzzer

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Size       int64
	Lines      int
	Words      int
	Duration   int64
}

func (r Result) String() string {
	return fmt.Sprintf("URL: %s | Status: %d | Size: %d bytes | Lines: %d | Words: %d | Time: %d ms",
		r.URL, r.StatusCode, r.Size, r.Lines, r.Words, r.Duration)
}

func (f *Fuzzer) worker(id int, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for url := range jobs {
		startTime := time.Now()
		resp, err := f.Client.Get(url)
		if err != nil {
			fmt.Printf("[Worker %d] Error occurred while fuzzing %s: %v\n", id, url, err)
			continue
		}
		duration := time.Since(startTime).Milliseconds()
		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		results <- Result{
			URL:        url,
			StatusCode: resp.StatusCode,
			Size:       resp.ContentLength,
			Lines:      bytes.Count(bodyBytes, []byte{'\n'}),
			Words:      len(bytes.Fields(bodyBytes)),
			Duration:   duration,
		}
	}

}
