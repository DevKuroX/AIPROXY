package filters

import (
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

func init() {
	Register(&smartTruncateFilter{})
}

type smartTruncateFilter struct{}

func (f *smartTruncateFilter) Name() string { return rtk.FilterSmartTruncate }

// ref: open-sse/rtk/filters/smartTruncate.js:5
func (f *smartTruncateFilter) Apply(input string) (string, error) {
	lines := strings.Split(input, "\n")
	if len(lines) < rtk.SmartTruncateMinLines {
		return input, nil
	}

	head := lines[:rtk.SmartTruncateHead]
	tail := lines[len(lines)-rtk.SmartTruncateTail:]
	cut := len(lines) - len(head) - len(tail)

	result := make([]string, 0, len(head)+1+len(tail))
	result = append(result, head...)
	result = append(result, "... +"+itoa(cut)+" lines truncated")
	result = append(result, tail...)

	return strings.Join(result, "\n"), nil
}
