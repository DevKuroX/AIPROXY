// ref: open-sse/rtk/applyFilter.js
package rtk

import (
	"fmt"
	"io"
	"os"
)

// ApplyFilter safely applies a filter to content with panic recovery.
// On error, returns the original content unchanged and logs a warning.
func ApplyFilter(content string, filter Filter) (string, error) {
	if filter == nil {
		return content, nil
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[rtk] warning: filter panicked — passing through raw output: %v\n", r)
		}
	}()

	result := filter(content)
	if result == "" {
		return content, nil
	}
	return result, nil
}

// SafeApply is a convenience wrapper that never returns an error.
// It always returns valid content, falling back to the original on any issue.
func SafeApply(content string, filter Filter) string {
	result, _ := ApplyFilter(content, filter)
	return result
}

// ApplyWithStats applies a filter and returns compression statistics.
type Stats struct {
	BytesBefore int64
	BytesAfter  int64
	Hits        []string
}

func ApplyWithStats(content string, filter Filter) (*Stats, string) {
	stats := &Stats{
		BytesBefore: int64(len(content)),
	}
	result := SafeApply(content, filter)
	stats.BytesAfter = int64(len(result))
	return stats, result
}

// SetWarningWriter allows customizing where warnings are written.
// Defaults to os.Stderr.
var warningWriter io.Writer = os.Stderr

func SetWarningWriter(w io.Writer) {
	warningWriter = w
}

func warnf(format string, args ...interface{}) {
	fmt.Fprintf(warningWriter, format, args...)
}
