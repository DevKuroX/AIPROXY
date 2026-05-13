// ref: open-sse/rtk/filters/grep.js
package filters

import (
	"sort"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

// GrepFilter compacts grep output.
// ref: open-sse/rtk/filters/grep.js:5
func GrepFilter(content string) string {
	byFile := make(map[string][][2]string)
	var total int

	for _, line := range strings.Split(content, "\n") {
		first := strings.Index(line, ":")
		if first == -1 {
			continue
		}
		second := strings.Index(line[first+1:], ":")
		if second == -1 {
			continue
		}
		second += first + 1

		file := line[:first]
		lineNumStr := line[first+1 : second]
		lineContent := line[second+1:]

		if !isDigits(lineNumStr) {
			continue
		}
		total++
		byFile[file] = append(byFile[file], [2]string{lineNumStr, lineContent})
	}

	if total == 0 {
		return content
	}

	files := make([]string, 0, len(byFile))
	for f := range byFile {
		files = append(files, f)
	}
	sort.Strings(files)

	var out strings.Builder
	out.WriteString(itoa(total) + " matches in " + itoa(len(files)) + "F:\n\n")

	for _, file := range files {
		matches := byFile[file]
		out.WriteString("[file] " + file + " (" + itoa(len(matches)) + "):\n")

		show := matches
		if len(matches) > rtk.GrepPerFileMax {
			show = matches[:rtk.GrepPerFileMax]
		}

		for _, m := range show {
			lineNum := m[0]
			lineContent := strings.TrimSpace(m[1])
			out.WriteString("  " + padLeft(lineNum, 4) + ": " + lineContent + "\n")
		}

		if len(matches) > rtk.GrepPerFileMax {
			out.WriteString("  +" + itoa(len(matches)-rtk.GrepPerFileMax) + "\n")
		}
		out.WriteString("\n")
	}

	return out.String()
}

func isDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func padLeft(s string, width int) string {
	for len(s) < width {
		s = " " + s
	}
	return s
}

func init() {
	rtk.RegisterFilter(rtk.FilterGrep, GrepFilter)
}
