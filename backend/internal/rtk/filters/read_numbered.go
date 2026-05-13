package filters

import (
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

func init() {
	Register(&readNumberedFilter{})
}

type readNumberedFilter struct{}

func (f *readNumberedFilter) Name() string { return rtk.FilterReadNumbered }

// ref: open-sse/rtk/filters/readNumbered.js:7
func (f *readNumberedFilter) Apply(input string) (string, error) {
	lines := strings.Split(input, "\n")
	if len(lines) < rtk.SmartTruncateMinLines {
		return input, nil
	}

	head := lines[:rtk.SmartTruncateHead]
	tail := lines[len(lines)-rtk.SmartTruncateTail:]
	cut := len(lines) - len(head) - len(tail)

	result := make([]string, 0, len(head)+1+len(tail))
	result = append(result, head...)
	result = append(result, "... +"+itoa(cut)+" lines truncated (file continues)")
	result = append(result, tail...)

	return strings.Join(result, "\n"), nil
}
