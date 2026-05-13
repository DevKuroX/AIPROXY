// ref: open-sse/rtk/filters/gitStatus.js
package filters

import (
	"regexp"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/rtk"
)

// GitStatusFilter compacts git status output.
// ref: open-sse/rtk/filters/gitStatus.js:13
func GitStatusFilter(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || (len(lines) == 1 && strings.TrimSpace(lines[0]) == "") {
		return "Clean working tree"
	}

	var branch string
	var stagedFiles, modifiedFiles, untrackedFiles []string
	var staged, modified, untracked, conflicts int

	longBranchRe := regexp.MustCompile(`^On branch (\S+)`)
	porcelainRe := regexp.MustCompile(`^[ MADRCU?!][ MADRCU?!] `)
	longFormRe := regexp.MustCompile(`^\s*(modified|new file|deleted|renamed|both modified):\s+(.+)$`)

	for _, raw := range lines {
		if strings.TrimSpace(raw) == "" {
			continue
		}

		// Long-form branch detection
		if m := longBranchRe.FindStringSubmatch(raw); m != nil {
			branch = m[1]
			continue
		}

		// Porcelain branch header
		if strings.HasPrefix(raw, "##") {
			branch = strings.TrimPrefix(raw, "##")
			branch = strings.TrimSpace(branch)
			continue
		}

		// Porcelain status
		if len(raw) >= 3 && porcelainRe.MatchString(raw) {
			x := raw[0]
			y := raw[1]
			file := raw[3:]

			if raw[:2] == "??" {
				untracked++
				untrackedFiles = append(untrackedFiles, file)
				continue
			}

			if strings.Contains("MADRC", string(x)) {
				staged++
				stagedFiles = append(stagedFiles, file)
			} else if x == 'U' {
				conflicts++
			}

			if y == 'M' || y == 'D' {
				modified++
				modifiedFiles = append(modifiedFiles, file)
			}
			continue
		}

		// Long form fallback
		if m := longFormRe.FindStringSubmatch(raw); m != nil {
			kind := m[1]
			path := strings.TrimSpace(m[2])
			switch kind {
			case "both modified":
				conflicts++
			case "modified", "deleted":
				modified++
				modifiedFiles = append(modifiedFiles, path)
			case "new file", "renamed":
				staged++
				stagedFiles = append(stagedFiles, path)
			}
		}
	}

	var out strings.Builder
	if branch != "" {
		out.WriteString("* " + branch + "\n")
	}

	if staged > 0 {
		out.WriteString("+ Staged: " + itoa(staged) + " files\n")
		for _, f := range sliceLimit(stagedFiles, rtk.StatusMaxFiles) {
			out.WriteString("   " + f + "\n")
		}
		if len(stagedFiles) > rtk.StatusMaxFiles {
			out.WriteString("   ... +" + itoa(len(stagedFiles)-rtk.StatusMaxFiles) + " more\n")
		}
	}

	if modified > 0 {
		out.WriteString("~ Modified: " + itoa(modified) + " files\n")
		for _, f := range sliceLimit(modifiedFiles, rtk.StatusMaxFiles) {
			out.WriteString("   " + f + "\n")
		}
		if len(modifiedFiles) > rtk.StatusMaxFiles {
			out.WriteString("   ... +" + itoa(len(modifiedFiles)-rtk.StatusMaxFiles) + " more\n")
		}
	}

	if untracked > 0 {
		out.WriteString("? Untracked: " + itoa(untracked) + " files\n")
		for _, f := range sliceLimit(untrackedFiles, rtk.StatusMaxUntracked) {
			out.WriteString("   " + f + "\n")
		}
		if len(untrackedFiles) > rtk.StatusMaxUntracked {
			out.WriteString("   ... +" + itoa(len(untrackedFiles)-rtk.StatusMaxUntracked) + " more\n")
		}
	}

	if conflicts > 0 {
		out.WriteString("conflicts: " + itoa(conflicts) + " files\n")
	}

	if staged == 0 && modified == 0 && untracked == 0 && conflicts == 0 {
		out.WriteString("clean — nothing to commit\n")
	}

	return strings.TrimRight(out.String(), "\n")
}

func init() {
	rtk.RegisterFilter(rtk.FilterGitStatus, GitStatusFilter)
}
