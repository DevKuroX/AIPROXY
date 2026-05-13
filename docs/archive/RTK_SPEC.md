
> **Setup-pack correction:** `git-log` (`FILTERS.GIT_LOG`) is declared in
> `open-sse/rtk/constants.js:45` but has **no implementation** in
> `open-sse/rtk/filters/` and is **not registered** in
> `open-sse/rtk/registry.js`. It's effectively dead code in 9router.
> **Do NOT port `git-log` as a working filter.** Mirror only the constant
> if you want, but skip §7.`git-log` below or mark it as no-op.

# docs/RTK_SPEC.md — RTK Token Saver (Parity Spec)

> Port of `_ref/9router/open-sse/rtk/`. Every behavior listed here is what the
> Go implementation MUST replicate to pass golden-fixture tests.

---

## 1. What RTK does

RTK = **R**equest **T**ool-result **K**ompressor.

It walks every message in a chat request body and, when it finds a tool-result
content block, attempts to detect its shape (git-diff, grep, ls, tree, etc.)
and applies a lossless or near-lossless compression filter to the text.

Goal: 20–40 % token reduction on coding sessions where tool outputs dominate
prompt size.

**RTK only mutates the *request* sent to the upstream provider.** The client
sees nothing. The model sees compressed content.

---

## 2. When RTK runs

In the request pipeline:

```
incoming body
   │
   ▼
[RTK.CompressMessages(body, enabled)]    ← THIS step
   │
   ▼
[translator.Convert(body, srcFormat, dstFormat)]
   │
   ▼
[caveman.Inject(body, dstFormat, level)]
   │
   ▼
[executor.Send(body)]
```

RTK runs **before** translation so the message shape it sees is always the
source-format shape (OpenAI chat, OpenAI Responses, or Claude messages).

---

## 3. Toggling

| Source | Effect |
|---|---|
| `settings.rtk_enabled` (global, in DB) | Default ON state |
| Request header `X-9R-RTK: off` | Disable for this request only |
| Request header `X-9R-RTK: on` | Force enable for this request only |

If disabled at any of these layers, RTK is a no-op (returns nil stats).

---

## 4. Message shapes it must handle

Reference: `_ref/9router/open-sse/rtk/index.js` lines 8–95.

### Shape 1 — OpenAI tool message, string content
```json
{ "role": "tool", "content": "<TEXT>" }
```
Compress `content` in place.

### Shape 1b — OpenAI tool message, array content
```json
{
  "role": "tool",
  "content": [{ "type": "text", "text": "<TEXT>" }]
}
```
For each part where `type == "text"`, compress `text`.

### Shape 2 — Claude tool_result, string content
Inside a message with `role == "user"` (or any role), `content` is an array
containing blocks like:
```json
{ "type": "tool_result", "tool_use_id": "...", "content": "<TEXT>" }
```
Compress `content`.

### Shape 3 — Claude tool_result, array content
```json
{
  "type": "tool_result",
  "tool_use_id": "...",
  "content": [{ "type": "text", "text": "<TEXT>" }]
}
```
For each part where `type == "text"`, compress `text`.

### Shape 4 — OpenAI Responses function_call_output
At the **top level** of the `input[]` array:
```json
{ "type": "function_call_output", "output": "<TEXT>" }
```
or array form:
```json
{
  "type": "function_call_output",
  "output": [{ "type": "input_text", "text": "<TEXT>" }]
}
```

### Important guards (per JS source)

- `block.is_error === true` → **skip** (preserve error traces).
- Container can be either `body.messages` (chat) or `body.input` (Responses).
- If neither is an array, return `nil`.

---

## 5. Per-text compression decision

In `compressText(text, stats, shape)`:

```text
let bytesIn = len(text)
if bytesIn < MIN_COMPRESS_SIZE { skip }      // too small to matter
if bytesIn > RAW_CAP            { skip }     // too huge — bail to avoid pathological cases
filter := autoDetectFilter(text)
if filter == nil                { skip }
out, ok := safeApply(filter, text)
if !ok                          { skip }
return out (with stats updated)
```

Constants (port `constants.js`):

```go
const (
    MIN_COMPRESS_SIZE = 512        // bytes
    RAW_CAP           = 5_000_000  // 5 MB hard upper bound

    // smart_truncate thresholds
    SMART_TRUNCATE_HEAD       = 120
    SMART_TRUNCATE_TAIL       = 60
    SMART_TRUNCATE_MIN_LINES  = 250

    // read_numbered detection
    READ_NUMBERED_MIN_HIT_RATIO = 0.7

    // search_list (Cursor file enumeration)
    SEARCH_LIST_TOTAL_DIR_MAX = 20
)
```

Filter names (must match these strings — they appear in stats/logs):

```go
const (
    FilterGitDiff       = "git-diff"
    FilterGitStatus     = "git-status"
    FilterGitLog        = "git-log"
    FilterGrep          = "grep"
    FilterFind          = "find"
    FilterLs            = "ls"
    FilterTree          = "tree"
    FilterDedupLog      = "dedup-log"
    FilterSmartTruncate = "smart-truncate"
    FilterReadNumbered  = "read-numbered"
    FilterSearchList    = "search-list"
)
```

---

## 6. Autodetect logic (port of `autodetect.js`)

Detection order matters — first match wins:

1. **git-diff** — first non-empty line starts with `diff --git ` OR `commit ` followed by hex.
2. **git-status** — first non-empty line matches `^On branch ` OR `Changes (not staged|to be committed):`
3. **git-log** — first non-empty line matches `^commit [a-f0-9]{7,}`
4. **grep** — ≥ 1 line of form `path:line:content` (≥3 colon-separated segments, segment 2 is digits).
5. **find** / **ls** / **tree** — path-like lines without colons.
   - **tree** if lines start with tree-drawing chars (`├──`, `└──`, `│  `).
   - **ls** if all path-like with no slashes (basename only).
   - **find** otherwise (path-like with slashes).
6. **read-numbered** — ≥ `MIN_LINES` (250) AND ≥ `READ_NUMBERED_MIN_HIT_RATIO` (0.7) of lines look like `  N|content`.
7. **dedup-log** — fallback when ≥ 5 non-empty lines and nothing else matched.
8. **smart-truncate** — last resort for big blobs (≥ `SMART_TRUNCATE_MIN_LINES` lines).
9. Else: `nil` (no compression).

`search-list` is *not* in autodetect — it's invoked when Cursor's
`list_dir` shape is recognized via shape inspection in `compress.go` (port
the JS pattern).

### Helpers (port verbatim)

```go
// isGrepLine: at least 3 colon-segments, segment 2 is purely digits
func isGrepLine(line string) bool { ... }

// isPathLike: non-empty, no ':', starts with '.', '/', or contains '/'
func isPathLike(line string) bool { ... }

// isLineNumbered: ≥ READ_NUMBERED_MIN_HIT_RATIO of non-empty lines match `^\s*\d+\s*\|`
func isLineNumbered(lines []string) bool { ... }
```

---

## 7. Per-filter behavior

For each filter: read the corresponding JS file in
`_ref/9router/open-sse/rtk/filters/` and port verbatim. Key behaviors:

### `git-diff`
- Group hunks by file.
- Truncate identical adjacent context lines to first/last 3.
- Collapse runs of unchanged lines longer than 6 to a `… (N lines unchanged) …` summary.
- Preserve all `+` / `-` lines.

### `git-status`
- Keep "On branch X" header.
- Group by section ("Changes not staged" / "Untracked").
- Drop "no changes added to commit" footer if present.
- Truncate file list >50 entries with `(+N more)`.

### `git-log`
- Keep first 20 commits in full, then `(+N more commits)`.

### `grep`
- Group consecutive matches in the same file with file-header line.
- Truncate runs >30 matches per file: keep first 20 + last 5, marker in middle.

### `find` / `ls`
- Sort alphabetically.
- Deduplicate.
- Truncate >100 entries: first 50, marker, last 25.

### `tree`
- Truncate sub-trees deeper than 3 levels with `… (N items hidden)`.

### `dedup-log`
- Collapse identical consecutive lines: `<line> (xN)`.
- Then collapse identical lines repeated >5 times across the text.

### `smart-truncate`
- Keep first `SMART_TRUNCATE_HEAD` (120) and last `SMART_TRUNCATE_TAIL` (60) lines.
- Insert middle marker: `… (N lines truncated) …`.

### `read-numbered`
- Detect Cursor's `read_file` output with `  N|content` lines.
- Apply `smart-truncate` but preserve line numbers in the marker:
  `… (lines A–B truncated) …`.

### `search-list`
- Cursor `list_dir` output. Show first `SEARCH_LIST_TOTAL_DIR_MAX` (20)
  directories in full, summarize the rest.

---

## 8. Safe-apply wrapper

Port `applyFilter.js`:

```go
func SafeApply(fn FilterFunc, text string) (out string, ok bool) {
    defer func() {
        if r := recover(); r != nil {
            slog.Warn("rtk: filter panicked, returning original",
                "panic", r)
            out, ok = text, false
        }
    }()
    out = fn(text)
    ok = true
    return
}
```

Rationale: if any filter has a bug, **never break the user's request** —
fall back to original text.

---

## 9. Stats output

`Compress` returns a `*Stats`:

```go
type Stats struct {
    BytesBefore int
    BytesAfter  int
    Hits        []Hit
}

type Hit struct {
    Shape  string  // "openai-tool", "claude-string", "openai-responses-array", ...
    Filter string  // "git-diff", "grep", ...
    Before int
    After  int
}
```

Returned to router. Logged at info level when `>0` hits. Persisted to usage
row if a save metric is requested.

---

## 10. Public Go API

```go
// internal/rtk/compress.go
package rtk

type Stats struct { ... }

// CompressMessages mutates body in place, returns stats (nil if disabled
// or body has no compressible content).
func CompressMessages(body any, enabled bool) *Stats
```

`body` is typed as `any` because at this point in the pipeline it's an
`map[string]any` parsed from JSON (we have not yet bound to a struct).
Casting/reflection inside `compress.go` mirrors the JS dynamic walk.

---

## 11. Testing strategy

### Golden fixtures

For every filter, capture from **real** outputs:

```
testdata/rtk/<filter>/<case>/
├── in.txt         (raw tool output)
└── out.txt        (expected compressed)
```

Test:
```go
func TestFilter_GitDiff(t *testing.T) {
    cases, _ := os.ReadDir("testdata/rtk/git-diff")
    for _, c := range cases {
        t.Run(c.Name(), func(t *testing.T) {
            in, _  := os.ReadFile(...)
            want, _ := os.ReadFile(...)
            got := filters.GitDiff(string(in))
            require.Equal(t, string(want), got)
        })
    }
}
```

### End-to-end fixtures

Capture full request bodies from 9router with RTK on:

```
testdata/rtk/e2e/<case>/
├── request.in.json    (raw client request)
├── request.out.json   (after RTK, before translation)
└── stats.json
```

Use `task parity:rtk` to run.

### Capture script

`scripts/capture_fixtures.sh`:
1. Boot 9router locally with debug log enabled.
2. Run a curated set of `curl` requests with known tool outputs.
3. Parse `~/.9router/log.txt` to extract before/after pairs.
4. Drop into `testdata/rtk/`.

---

## 12. Things NOT in scope for RTK

- Compressing **response** content — out of scope (model output is what the
  user wants verbatim).
- Compressing user-authored messages — only tool outputs.
- Lossy summarization — every filter is reversible-enough for the model
  to still answer correctly. We never drop information the model needs to
  reason about (we drop redundancy).

---

## 13. Reference

| Concept | 9router file |
|---|---|
| Entry / shape walk | `open-sse/rtk/index.js` |
| Constants | `open-sse/rtk/constants.js` |
| Autodetect | `open-sse/rtk/autodetect.js` |
| Safe apply | `open-sse/rtk/applyFilter.js` |
| Filter registry | `open-sse/rtk/registry.js` |
| Filters | `open-sse/rtk/filters/*.js` |
| Caveman (separate) | `open-sse/rtk/caveman.js`, `cavemanPrompts.js` |
