package fuzzer

import (
	"bytes"
	"fmirage/internal/output"
	"fmt"
	"io"
	"sync"
	"time"
)

func (f *Fuzzer) worker(id int, jobs <-chan string, results chan<- output.Result, wg *sync.WaitGroup) {
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
		if err != nil {
			fmt.Printf("[Worker %d] Error reading response body for %s: %v\n", id, url, err)
			continue
		}

		size := resp.ContentLength
		if size < 0 {
			size = int64(len(bodyBytes))
		}

		result := output.Result{
			URL:        url,
			StatusCode: resp.StatusCode,
			Size:       size,
			Lines:      bytes.Count(bodyBytes, []byte{'\n'}),
			Words:      len(bytes.Fields(bodyBytes)),
			Duration:   duration,
		}

		if (!f.Cfg.DisableMatcher && !f.Filter.ShouldMatch(&result)) ||
			f.Filter.ShouldDrop(&result) {
			continue
		}

		results <- result
	}

}
