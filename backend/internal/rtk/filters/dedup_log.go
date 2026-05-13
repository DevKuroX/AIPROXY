package filters

import (
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

func init() {
	Register(&dedupLogFilter{})
}

type dedupLogFilter struct{}

func (f *dedupLogFilter) Name() string { return rtk.FilterDedupLog }

// ref: open-sse/rtk/filters/dedupLog.js:4
func (f *dedupLogFilter) Apply(input string) (string, error) {
	lines := strings.Split(input, "\n")
	var out []string
	var prev string
	runCount := 0
	blankStreak := 0

	flushRun := func() {
		if prev != "" && runCount > 1 {
			out = append(out, "  ... ("+itoa(runCount-1)+" duplicate lines)")
		}
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if blankStreak < 1 {
				out = append(out, line)
			}
			blankStreak++
			flushRun()
			prev = ""
			runCount = 0
			continue
		}
		blankStreak = 0
		if line == prev {
			runCount++
			continue
		}
		flushRun()
		out = append(out, line)
		prev = line
		runCount = 1
		if len(out) >= rtk.DedupLineMax {
			out = append(out, "... (truncated at "+itoa(rtk.DedupLineMax)+" lines)")
			return strings.Join(out, "\n"), nil
		}
	}
	flushRun()
	return strings.Join(out, "\n"), nil
}
