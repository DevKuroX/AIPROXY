// Package rtk provides content compression and filtering for LLM request bodies.
// ref: open-sse/rtk/constants.js
package rtk

// Size thresholds (mirror Rust defaults)
const (
	RawCap             = 10 * 1024 * 1024 // 10 MiB
	MinCompressSize    = 500              // bytes; skip tiny blobs
	DetectWindow       = 1024             // autodetect peeks first N chars
	GitDiffHunkMaxLines = 100             // per-hunk line cap
	GitDiffContextKeep = 3                // context lines around changes
	DedupLineMax       = 2000             // dedupLog truncation cap
)

// Rust pipe_cmd.rs parity caps
const (
	GrepPerFileMax     = 10 // match rust: matches.iter().take(10)
	FindPerDirMax      = 10 // match rust: files.iter().take(10)
	FindTotalDirMax    = 20 // match rust: dirs.iter().take(20)
)

// git status caps (rust config::limits())
const (
	StatusMaxFiles       = 10 // config::limits().status_max_files
	StatusMaxUntracked   = 10 // config::limits().status_max_untracked
)

// ls compact_ls (rtk/src/cmds/system/ls.rs)
const (
	LsExtSummaryTop = 5 // top-N extensions in summary
)

// LSNoiseDirs are directories to skip in ls output
var LSNoiseDirs = []string{
	"node_modules", ".git", "target", "__pycache__",
	".next", "dist", "build", ".venv", "venv",
	".cache", ".idea", ".vscode", ".DS_Store",
}

// tree filter_tree_output cap (no rust cap, we add one to be safe)
const TreeMaxLines = 200

// Cursor Glob "Result of search in '...' (total N files):" list
const (
	SearchListPerDirMax   = 10
	SearchListTotalDirMax = 20
)

// Smart truncate (port of filter.rs smart_truncate fallback)
const (
	SmartTruncateHead    = 120 // lines kept from top
	SmartTruncateTail    = 60  // lines kept from bottom
	SmartTruncateMinLines = 250 // only kick in above this
)

// readNumbered (files with "  N|content" lines, e.g. Cursor read_file)
const ReadNumberedMinHitRatio = 0.7

// Filter name strings (Rust parity + JS extras)
// Note: git-log is NOT included (not implemented in 9router registry)
const (
	FilterGitDiff       = "git-diff"
	FilterGitStatus     = "git-status"
	FilterGrep          = "grep"
	FilterFind          = "find"
	FilterLs            = "ls"
	FilterTree          = "tree"
	FilterDedupLog      = "dedup-log"
	FilterSmartTruncate = "smart-truncate"
	FilterReadNumbered  = "read-numbered"
	FilterSearchList    = "search-list"
)
