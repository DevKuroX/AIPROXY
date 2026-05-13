// ref: open-sse/rtk/autodetect.js
package rtk

import (
	"regexp"
	"strings"
)

// Content type detection patterns
var (
	reGitDiff      = regexp.MustCompile(`(?m)^diff --git `)
	reGitDiffHunk  = regexp.MustCompile(`(?m)^@@ `)
	reGitStatus    = regexp.MustCompile(`(?m)^On branch |^nothing to commit|^Changes (not |to be )|^Untracked files:`)
	rePorcelain    = regexp.MustCompile(`(?m)^[ MADRCU?!][ MADRCU?!] \S`)
	reTreeGlyph    = regexp.MustCompile(`[├└]──|│  `)
	reLsRow        = regexp.MustCompile(`(?m)^[-dlbcps][rwx-]{9}`)
	reLsTotal      = regexp.MustCompile(`(?m)^total \d+$`)
	reGrepLine     = regexp.MustCompile(`^[^:]+:\d+:`)
	reSearchList   = regexp.MustCompile(`(?m)^Result of search in '.*' \(total \d+ files?\):`)
	reReadNumbered = regexp.MustCompile(`^\s*\d+\|`)
)

// DetectContentType analyzes content and returns the detected filter name.
// ref: open-sse/rtk/autodetect.js:24
func DetectContentType(content string) string {
	if content == "" {
		return ""
	}

	head := content
	if len(content) > DetectWindow {
		head = content[:DetectWindow]
	}

	if reGitDiff.MatchString(head) || reGitDiffHunk.MatchString(head) {
		return FilterGitDiff
	}

	if reGitStatus.MatchString(head) || isMostlyPorcelain(head) {
		return FilterGitStatus
	}

	lines := strings.Split(head, "\n")
	var nonEmpty []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			nonEmpty = append(nonEmpty, l)
		}
	}

	// Rust grep rule: first 5 non-empty lines, ANY matches "file:number:content"
	first5 := nonEmpty
	if len(first5) > 5 {
		first5 = first5[:5]
	}
	for _, l := range first5 {
		if isGrepLine(l) {
			return FilterGrep
		}
	}

	// Rust find rule: ALL non-empty lines path-like (no ':'), >=3 lines
	if len(nonEmpty) >= 3 {
		allPathLike := true
		for _, l := range nonEmpty {
			if !isPathLike(l) {
				allPathLike = false
				break
			}
		}
		if allPathLike {
			return FilterFind
		}
	}

	// Tree: contains box-drawing glyphs typical of `tree` command
	if reTreeGlyph.MatchString(head) {
		return FilterTree
	}

	// ls -la: has "total N" header or >=3 rows starting with perms string
	if reLsTotal.MatchString(head) || countMatches(head, reLsRow) >= 3 {
		return FilterLs
	}

	// Cursor Glob search list header
	if reSearchList.MatchString(head) {
		return FilterSearchList
	}

	// Line-numbered file dump ("  N|content") — fire only if many lines match
	if len(lines) >= SmartTruncateMinLines && isLineNumbered(lines) {
		return FilterReadNumbered
	}

	// Fallback: dedupLog for generic multi-line noise with duplicates
	if len(nonEmpty) >= 5 {
		return FilterDedupLog
	}

	// Last resort: big blob with no structure — smart truncate
	if len(content) >= MinCompressSize {
		return FilterSmartTruncate
	}

	return ""
}

// SelectFilter returns the filter for the given content type.
func SelectFilter(contentType string) Filter {
	if contentType == "" {
		return nil
	}
	f, ok := GetFilter(contentType)
	if !ok {
		return nil
	}
	return f
}

// AutoSelectFilter detects content type and returns the appropriate filter.
func AutoSelectFilter(content string) Filter {
	return SelectFilter(DetectContentType(content))
}

func isMostlyPorcelain(head string) bool {
	lines := strings.Split(head, "\n")
	if len(lines) < 3 {
		return false
	}
	matches := 0
	for _, l := range lines {
		if rePorcelain.MatchString(l) {
			matches++
		}
	}
	return matches >= len(lines)/2
}

func isGrepLine(l string) bool {
	return reGrepLine.MatchString(l)
}

func isPathLike(l string) bool {
	return !strings.Contains(l, ":") && strings.TrimSpace(l) != ""
}

func countMatches(s string, re *regexp.Regexp) int {
	matches := re.FindAllStringIndex(s, -1)
	return len(matches)
}

func isLineNumbered(lines []string) bool {
	if len(lines) == 0 {
		return false
	}
	matches := 0
	for _, l := range lines {
		if reReadNumbered.MatchString(l) {
			matches++
		}
	}
	ratio := float64(matches) / float64(len(lines))
	return ratio >= ReadNumberedMinHitRatio
}
