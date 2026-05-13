// Package filters provides content compression filters for RTK.
// ref: open-sse/rtk/filters/gitDiff.js
package filters

import (
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

// GitDiffFilter compacts unified diff output.
// ref: open-sse/rtk/filters/gitDiff.js:5-90
func GitDiffFilter(content string) string {
	return gitDiff(content, 500)
}

// gitDiff is the core implementation with configurable max lines.
// ref: open-sse/rtk/filters/gitDiff.js:5
func gitDiff(diff string, maxLines int) string {
	var result []string
	var currentFile string
	var added, removed int
	var inHunk bool
	var hunkShown, hunkSkipped int
	var wasTruncated bool
	maxHunkLines := rtk.GitDiffHunkMaxLines

	lines := strings.Split(diff, "\n")

outer:
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if hunkSkipped > 0 {
				result = append(result, "  ... ("+itoa(hunkSkipped)+" lines truncated)")
				wasTruncated = true
				hunkSkipped = 0
			}
			if currentFile != "" && (added > 0 || removed > 0) {
				result = append(result, "  +"+itoa(added)+" -"+itoa(removed))
			}
			parts := strings.SplitN(line, " b/", 2)
			if len(parts) > 1 {
				currentFile = parts[1]
			} else {
				currentFile = "unknown"
			}
			result = append(result, "\n"+currentFile)
			added = 0
			removed = 0
			inHunk = false
			hunkShown = 0
		} else if strings.HasPrefix(line, "@@") {
			if hunkSkipped > 0 {
				result = append(result, "  ... ("+itoa(hunkSkipped)+" lines truncated)")
				wasTruncated = true
				hunkSkipped = 0
			}
			inHunk = true
			hunkShown = 0
			result = append(result, "  "+line)
		} else if inHunk {
			if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
				added++
				if hunkShown < maxHunkLines {
					result = append(result, "  "+line)
					hunkShown++
				} else {
					hunkSkipped++
				}
			} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
				removed++
				if hunkShown < maxHunkLines {
					result = append(result, "  "+line)
					hunkShown++
				} else {
					hunkSkipped++
				}
			} else if hunkShown < maxHunkLines && !strings.HasPrefix(line, "\\") {
				if hunkShown > 0 {
					result = append(result, "  "+line)
					hunkShown++
				}
			}
		}

		if len(result) >= maxLines {
			result = append(result, "\n... (more changes truncated)")
			wasTruncated = true
			break outer
		}
	}

	if hunkSkipped > 0 {
		result = append(result, "  ... ("+itoa(hunkSkipped)+" lines truncated)")
		wasTruncated = true
	}

	if currentFile != "" && (added > 0 || removed > 0) {
		result = append(result, "  +"+itoa(added)+" -"+itoa(removed))
	}

	if wasTruncated {
		result = append(result, "[full diff: rtk git diff --no-compact]")
	}

	return strings.Join(result, "\n")
}

func sliceLimit[T any](s []T, max int) []T {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// itoa is a simple int to string conversion helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var neg bool
	if n < 0 {
		neg = true
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		return "-" + string(digits)
	}
	return string(digits)
}

func init() {
	rtk.RegisterFilter(rtk.FilterGitDiff, GitDiffFilter)
}
