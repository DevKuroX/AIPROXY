// ref: open-sse/rtk/filters/find.js
package filters

import (
	"sort"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

// FindFilter compacts find output.
// ref: open-sse/rtk/filters/find.js:5
func FindFilter(content string) string {
	lines := filterEmpty(strings.Split(content, "\n"))
	if len(lines) == 0 {
		return content
	}

	byDir := make(map[string][]string)

	for _, path := range lines {
		lastSlash := strings.LastIndex(path, "/")
		var dir, basename string
		if lastSlash == -1 {
			dir = "."
			basename = path
		} else {
			dir = path[:lastSlash]
			if dir == "" {
				dir = "/"
			}
			basename = path[lastSlash+1:]
		}
		byDir[dir] = append(byDir[dir], basename)
	}

	dirs := make([]string, 0, len(byDir))
	for d := range byDir {
		dirs = append(dirs, d)
	}
	sort.Strings(dirs)

	var out strings.Builder
	out.WriteString(itoa(len(lines)) + " files in " + itoa(len(dirs)) + " dirs:\n\n")

	showDirs := dirs
	if len(dirs) > rtk.FindTotalDirMax {
		showDirs = dirs[:rtk.FindTotalDirMax]
	}

	for _, dir := range showDirs {
		files := byDir[dir]
		out.WriteString(dir + "/ (" + itoa(len(files)) + "):\n")

		showFiles := files
		if len(files) > rtk.FindPerDirMax {
			showFiles = files[:rtk.FindPerDirMax]
		}

		for _, f := range showFiles {
			out.WriteString("  " + f + "\n")
		}

		if len(files) > rtk.FindPerDirMax {
			out.WriteString("  +" + itoa(len(files)-rtk.FindPerDirMax) + "\n")
		}
		out.WriteString("\n")
	}

	if len(dirs) > rtk.FindTotalDirMax {
		out.WriteString("+" + itoa(len(dirs)-rtk.FindTotalDirMax) + " more dirs\n")
	}

	return out.String()
}

func filterEmpty(lines []string) []string {
	var result []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			result = append(result, l)
		}
	}
	return result
}

func init() {
	rtk.RegisterFilter(rtk.FilterFind, FindFilter)
}
