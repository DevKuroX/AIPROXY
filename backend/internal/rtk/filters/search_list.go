package filters

import (
	"sort"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

func init() {
	Register(&searchListFilter{})
}

type searchListFilter struct{}

func (f *searchListFilter) Name() string { return rtk.FilterSearchList }

var searchListHeaderRe = regexpCompile(`^Result of search in '[^']*' \(total (\d+) files?\):`)

// ref: open-sse/rtk/filters/searchList.js:7
func (f *searchListFilter) Apply(input string) (string, error) {
	lines := strings.Split(input, "\n")
	if len(lines) == 0 {
		return input, nil
	}

	header := lines[0]
	rest := lines[1:]

	var paths []string
	for _, raw := range rest {
		t := strings.TrimSpace(raw)
		if !strings.HasPrefix(t, "- ") {
			continue
		}
		paths = append(paths, t[2:])
	}
	if len(paths) == 0 {
		return input, nil
	}

	byDir := make(map[string][]string)
	for _, p := range paths {
		slash := strings.LastIndex(p, "/")
		var dir, name string
		if slash == -1 {
			dir = "."
			name = p
		} else {
			dir = p[:slash]
			if dir == "" {
				dir = "/"
			}
			name = p[slash+1:]
		}
		byDir[dir] = append(byDir[dir], name)
	}

	dirs := make([]string, 0, len(byDir))
	for d := range byDir {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	var sb strings.Builder
	sb.WriteString(header)
	sb.WriteByte('\n')
	sb.WriteString(itoa(len(paths)))
	sb.WriteString(" files in ")
	sb.WriteString(itoa(len(dirs)))
	sb.WriteString(" dirs:\n\n")

	maxDirs := rtk.SearchListTotalDirMax
	if len(dirs) < maxDirs {
		maxDirs = len(dirs)
	}

	for _, dir := range dirs[:maxDirs] {
		names := byDir[dir]
		sb.WriteString(dir)
		sb.WriteString("/ (")
		sb.WriteString(itoa(len(names)))
		sb.WriteString("):\n")

		maxNames := rtk.SearchListPerDirMax
		if len(names) < maxNames {
			maxNames = len(names)
		}

		for _, n := range names[:maxNames] {
			sb.WriteString("  ")
			sb.WriteString(n)
			sb.WriteByte('\n')
		}
		if len(names) > rtk.SearchListPerDirMax {
			sb.WriteString("  +")
			sb.WriteString(itoa(len(names) - rtk.SearchListPerDirMax))
			sb.WriteByte('\n')
		}
		sb.WriteByte('\n')
	}

	if len(dirs) > rtk.SearchListTotalDirMax {
		sb.WriteString("+")
		sb.WriteString(itoa(len(dirs) - rtk.SearchListTotalDirMax))
		sb.WriteString(" more dirs\n")
	}

	return strings.TrimRight(sb.String(), "\n"), nil
}

func regexpCompile(pattern string) interface{ MatchString(string) bool } {
	return &simpleRegex{pattern: pattern}
}

type simpleRegex struct {
	pattern string
}

func (r *simpleRegex) MatchString(s string) bool {
	return len(s) >= len(r.pattern) && s[:len(r.pattern)] == r.pattern
}
