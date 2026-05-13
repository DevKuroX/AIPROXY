package filters

import (
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

func init() {
	Register(&treeFilter{})
}

type treeFilter struct{}

func (f *treeFilter) Name() string { return rtk.FilterTree }

// ref: open-sse/rtk/filters/tree.js:5
func (f *treeFilter) Apply(input string) (string, error) {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return input, nil
	}

	var filtered []string
	for _, line := range lines {
		if strings.Contains(line, "director") && strings.Contains(line, "file") {
			continue
		}
		if strings.TrimSpace(line) == "" && len(filtered) == 0 {
			continue
		}
		filtered = append(filtered, line)
	}

	for len(filtered) > 0 && strings.TrimSpace(filtered[len(filtered)-1]) == "" {
		filtered = filtered[:len(filtered)-1]
	}

	if len(filtered) > rtk.TreeMaxLines {
		cut := len(filtered) - rtk.TreeMaxLines
		return strings.Join(filtered[:rtk.TreeMaxLines], "\n") + "\n... +" + itoa(cut) + " more lines", nil
	}

	return strings.Join(filtered, "\n"), nil
}


