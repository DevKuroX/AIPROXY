// ref: open-sse/rtk/filters/ls.js
package filters

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

var lsDateRe = regexp.MustCompile(`\s+(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)\s+\d{1,2}\s+(\d{4}|\d{2}:\d{2})\s+`)

type lsParsedLine struct {
	fileType string
	size     int
	name     string
}

// LsFilter compacts ls -la output.
// ref: open-sse/rtk/filters/ls.js:34
func LsFilter(content string) string {
	var dirs []string
	var files [][2]string
	byExt := make(map[string]int)

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "total ") || len(line) == 0 {
			continue
		}
		parsed := parseLsLine(line)
		if parsed == nil {
			continue
		}
		if parsed.name == "." || parsed.name == ".." {
			continue
		}
		if containsStr(rtk.LSNoiseDirs, parsed.name) {
			continue
		}

		if parsed.fileType == "d" {
			dirs = append(dirs, parsed.name)
		} else if parsed.fileType == "-" || parsed.fileType == "l" {
			dot := strings.LastIndex(parsed.name, ".")
			var ext string
			if dot > 0 {
				ext = parsed.name[dot:]
			} else {
				ext = "no ext"
			}
			byExt[ext]++
			files = append(files, [2]string{parsed.name, humanSize(parsed.size)})
		}
	}

	if len(dirs) == 0 && len(files) == 0 {
		return content
	}

	var out strings.Builder
	for _, d := range dirs {
		out.WriteString(d + "/\n")
	}
	for _, f := range files {
		out.WriteString(f[0] + "  " + f[1] + "\n")
	}

	out.WriteString("\nSummary: " + itoa(len(files)) + " files, " + itoa(len(dirs)) + " dirs")
	if len(byExt) > 0 {
		ext := make([][2]string, 0, len(byExt))
		for e, c := range byExt {
			ext = append(ext, [2]string{e, itoa(c)})
		}
		sort.Slice(ext, func(i, j int) bool {
			ci, _ := strconv.Atoi(ext[i][1])
			cj, _ := strconv.Atoi(ext[j][1])
			return ci > cj
		})

		top := ext
		if len(ext) > rtk.LsExtSummaryTop {
			top = ext[:rtk.LsExtSummaryTop]
		}

		var parts []string
		for _, e := range top {
			parts = append(parts, e[1]+" "+e[0])
		}
		out.WriteString(" (" + strings.Join(parts, ", "))
		if len(ext) > rtk.LsExtSummaryTop {
			out.WriteString(", +" + itoa(len(ext)-rtk.LsExtSummaryTop) + " more")
		}
		out.WriteString(")")
	}

	return out.String()
}

func parseLsLine(line string) *lsParsedLine {
	m := lsDateRe.FindStringIndex(line)
	if m == nil {
		return nil
	}

	name := line[m[1]:]
	beforeDate := line[:m[0]]
	beforeParts := strings.Fields(beforeDate)
	if len(beforeParts) < 4 {
		return nil
	}

	perms := beforeParts[0]
	fileType := string(perms[0])

	size := 0
	for i := len(beforeParts) - 1; i >= 0; i-- {
		n, err := strconv.Atoi(beforeParts[i])
		if err == nil && strconv.Itoa(n) == beforeParts[i] {
			size = n
			break
		}
	}

	return &lsParsedLine{fileType: fileType, size: size, name: name}
}

func humanSize(bytes int) string {
	if bytes >= 1048576 {
		return strconv.FormatFloat(float64(bytes)/1048576, 'f', 1, 64) + "M"
	}
	if bytes >= 1024 {
		return strconv.FormatFloat(float64(bytes)/1024, 'f', 1, 64) + "K"
	}
	return itoa(bytes) + "B"
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func init() {
	rtk.RegisterFilter(rtk.FilterLs, LsFilter)
}
