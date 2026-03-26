package output

import (
	"fmt"
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
