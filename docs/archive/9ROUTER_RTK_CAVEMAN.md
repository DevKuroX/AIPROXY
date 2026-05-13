# 9Router RTK & Caveman Mode - Technical Deep Dive

**Reference**: `open-sse/rtk/` in 9router  
**Purpose**: Token compression for tool outputs (20-40% savings)

---

## RTK (Rust Token Killer) - Token Compression Engine

### Overview

RTK is a **lossless compression system** for tool outputs that runs **before format translation**. It detects structured tool output (git diff, grep, ls, etc.) and applies intelligent filters to reduce token count without losing information.

**Key Properties**:
- **Automatic**: No configuration needed
- **Safe**: If filter fails or makes output bigger, keeps original
- **Universal**: Works across all formats (0penAI, CL4ude, Gemini, etc.)
- **Default ON**: Toggle in Dashboard → Endpoint Settings

### Compression Results

**Real-World Example**:
```
Without RTK: 47,000 tokens sent to LLM
With RTK:    28,000 tokens sent to LLM
Savings:     40% (same context, same answer)
```

**Typical Savings by Tool**:
- `git diff`: 30-50% reduction
- `grep -rn`: 25-40% reduction
- `find`: 20-35% reduction
- `ls -la`: 15-30% reduction
- `tree`: 25-45% reduction
- Log dumps: 30-60% reduction

---

## Architecture

### Data Flow

```
1. REQUEST ARRIVES
   ↓
2. DETECT TOOL OUTPUT FORMAT
   ├─ Read first 1KB of tool_result
   ├─ Match against known patterns
   ├─ Return filter name or null
   ↓
3. APPLY FILTER (if detected)
   ├─ Run filter function
   ├─ Measure output size
   ├─ If bigger or error: keep original
   ├─ If smaller: use compressed
   ↓
4. LOG COMPRESSION STATS
   ├─ Bytes before/after
   ├─ Filter name
   ├─ Savings percentage
   ↓
5. CONTINUE TO FORMAT TRANSLATION
```

### File Structure

```
open-sse/rtk/
├── index.js                      # Main compression engine
├── autodetect.js                 # Format detection (1KB pattern matching)
├── applyFilter.js                # Safe filter application
├── caveman.js                    # Aggressive compression mode
├── cavemanPrompts.js             # Caveman prompt templates
├── constants.js                  # RTK constants (RAW_CAP, MIN_COMPRESS_SIZE)
├── registry.js                   # Filter registry
└── filters/                      # Compression filters
    ├── git-diff.js               # Remove redundant hunk headers
    ├── git-status.js             # Remove file paths
    ├── grep.js                   # Remove redundant file paths
    ├── find.js                   # Remove redundant directory prefixes
    ├── ls.js                     # Remove redundant directory info
    ├── tree.js                   # Collapse deep nesting
    ├── dedup-log.js              # Remove duplicate log lines
    ├── smart-truncate.js         # Intelligent truncation
    ├── read-numbered.js          # Remove line numbers
    └── search-list.js            # Deduplicate search results
```

---

## Core Functions

### `compressMessages(body, enabled)`

**Location**: `open-sse/rtk/index.js:10-60`

**Purpose**: Compress tool_result content in request body.

**Signature**:
```javascript
function compressMessages(body, enabled) {
  // Returns: { bytesBefore, bytesAfter, hits: [...] } or null
}
```

**Supported Message Shapes**:

1. **0penAI Tool Message** (string):
   ```javascript
   { role: "tool", content: "git diff output..." }
   ```

2. **0penAI Tool Message** (array):
   ```javascript
   { role: "tool", content: [{ type: "text", text: "..." }] }
   ```

3. **CL4ude Tool Message**:
   ```javascript
   { role: "user", content: [{ type: "tool_result", content: "..." }] }
   ```

4. **0penAI Responses** (string):
   ```javascript
   { type: "function_call_output", output: "..." }
   ```

5. **0penAI Responses** (array):
   ```javascript
   { type: "function_call_output", output: [{ type: "input_text", text: "..." }] }
   ```

**Algorithm**:
```javascript
for each message in body.messages or body.input:
  if message is tool_result:
    content = message.content or message.output
    if content is string:
      content = compressText(content)
    else if content is array:
      for each part in content:
        if part.type === "text":
          part.text = compressText(part.text)
```

### `compressText(text, stats, shape)`

**Location**: `open-sse/rtk/index.js:70-120`

**Purpose**: Compress single text string.

**Algorithm**:
```javascript
1. Measure original size
   bytes_before = text.length

2. Auto-detect format
   filter = autoDetectFilter(text)
   if no filter detected:
     return original text

3. Apply filter
   try:
     compressed = filter(text)
   catch error:
     return original text

4. Check if worth compressing
   if compressed.length >= bytes_before:
     return original text

5. Record stats
   stats.bytesBefore += bytes_before
   stats.bytesAfter += compressed.length
   stats.hits.push({ filter, bytes_before, bytes_after })

6. Return compressed
   return compressed
```

### `autoDetectFilter(text)`

**Location**: `open-sse/rtk/autodetect.js`

**Purpose**: Detect tool output format from first 1KB.

**Detection Patterns**:

| Pattern | Filter | Confidence |
|---------|--------|------------|
| `^diff --git` | `git-diff` | 100% |
| `^On branch` | `git-status` | 100% |
| `^[^:]+:[0-9]+:` | `grep` | 95% |
| `^[^:]+$` (many lines) | `find` | 90% |
| `^total [0-9]+` | `ls` | 95% |
| `^.` (tree-like) | `tree` | 85% |
| Duplicate lines | `dedup-log` | 80% |

**Return Value**:
```javascript
{
  name: "git-diff",           // Filter name
  confidence: 0.95,           // 0-1 confidence score
  filter: function(text) {}   // Filter function
}
```

---

## Compression Filters

### 1. git-diff Filter

**Purpose**: Remove redundant hunk headers and context.

**Input**:
```
diff --git a/file.js b/file.js
index abc123..def456 100644
--- a/file.js
+++ b/file.js
@@ -10,5 +10,6 @@
 function foo() {
   console.log("hello");
+  console.log("world");
 }
```

**Output**:
```
function foo() {
  console.log("hello");
+ console.log("world");
}
```

**Savings**: 30-50% (removes file paths, hunk headers, context lines)

### 2. grep Filter

**Purpose**: Remove redundant file paths.

**Input**:
```
src/index.js:10:function foo() {
src/index.js:11:  console.log("hello");
src/utils.js:5:function bar() {
src/utils.js:6:  return 42;
```

**Output**:
```
src/index.js:
  10: function foo() {
  11:   console.log("hello");
src/utils.js:
  5: function bar() {
  6:   return 42;
```

**Savings**: 25-40% (deduplicates file paths)

### 3. find Filter

**Purpose**: Remove redundant directory prefixes.

**Input**:
```
./src/components/Button.js
./src/components/Input.js
./src/utils/helpers.js
./src/utils/constants.js
```

**Output**:
```
src/
  components/Button.js
  components/Input.js
  utils/helpers.js
  utils/constants.js
```

**Savings**: 20-35% (collapses directory structure)

### 4. ls Filter

**Purpose**: Remove redundant directory info.

**Input**:
```
total 48
drwxr-xr-x  5 user  group   160 May 11 10:30 .
drwxr-xr-x 10 user  group   320 May 10 15:45 ..
-rw-r--r--  1 user  group  1234 May 11 10:30 file1.js
-rw-r--r--  1 user  group  5678 May 11 10:25 file2.js
```

**Output**:
```
file1.js (1.2K)
file2.js (5.5K)
```

**Savings**: 15-30% (removes permissions, owner, timestamps)

### 5. tree Filter

**Purpose**: Collapse deep nesting.

**Input**:
```
.
├── src/
│   ├── components/
│   │   ├── Button.js
│   │   ├── Input.js
│   │   └── Modal.js
│   ├── utils/
│   │   ├── helpers.js
│   │   └── constants.js
│   └── index.js
└── package.json
```

**Output**:
```
src/
  components/ (3 files)
  utils/ (2 files)
  index.js
package.json
```

**Savings**: 25-45% (collapses deep directories)

### 6. dedup-log Filter

**Purpose**: Remove duplicate log lines.

**Input**:
```
[INFO] Starting server
[INFO] Loading config
[INFO] Starting server
[INFO] Loading config
[INFO] Starting server
[ERROR] Connection failed
```

**Output**:
```
[INFO] Starting server (×3)
[INFO] Loading config (×2)
[ERROR] Connection failed
```

**Savings**: 30-60% (deduplicates repeated lines)

---

## Caveman Mode - Aggressive Compression

### Overview

**Caveman Mode** is an **optional, aggressive compression** that rewrites output in simplified "caveman-speak" to reduce output tokens by up to 65%.

**Example**:
```
Normal: "File write operation completed successfully"
Caveman: "file write ok"

Normal: "Connection to database established"
Caveman: "db connect ok"
```

**Trade-offs**:
- ✅ Saves 20-65% output tokens
- ❌ May reduce quality for strict system prompts
- ❌ Not recommended for critical tasks

**Default**: OFF (toggle in Dashboard)

### Implementation

**Location**: `open-sse/rtk/caveman.js`

**Function**:
```javascript
function injectCaveman(body, level) {
  // level: "low" | "medium" | "high"
  // Rewrites system prompt + tool outputs
}
```

**Caveman Prompts**: `open-sse/rtk/cavemanPrompts.js`

**Levels**:
- **Low**: Minimal rewriting, ~20% savings
- **Medium**: Moderate rewriting, ~40% savings
- **High**: Aggressive rewriting, ~65% savings

### Caveman Prompt Injection

**System Prompt Modification**:
```javascript
// Original
"You are a helpful coding assistant."

// Caveman (High)
"You are caveman. Speak simple. Short words. No fancy. Just facts."
```

**Tool Output Rewriting**:
```javascript
// Original
"Successfully created file src/index.js with 150 lines of code"

// Caveman
"file create ok: src/index.js (150 lines)"
```

---

## Configuration

### Constants

**Location**: `open-sse/rtk/constants.js`

```javascript
// Maximum raw output size to compress
export const RAW_CAP = 100_000;  // 100KB

// Minimum size to attempt compression
export const MIN_COMPRESS_SIZE = 500;  // 500 bytes

// Compression threshold
export const COMPRESSION_THRESHOLD = 0.95;  // Keep if >95% of original
```

### Filter Registry

**Location**: `open-sse/rtk/registry.js`

```javascript
export const FILTER_REGISTRY = {
  "git-diff": gitDiffFilter,
  "git-status": gitStatusFilter,
  "grep": grepFilter,
  "find": findFilter,
  "ls": lsFilter,
  "tree": treeFilter,
  "dedup-log": dedupLogFilter,
  "smart-truncate": smartTruncateFilter,
  "read-numbered": readNumberedFilter,
  "search-list": searchListFilter,
};
```

---

## Integration Points

### In chatCore.js

**Location**: `open-sse/handlers/chatCore.js:20-30`

```javascript
import { compressMessages, formatRtkLog } from "../rtk/index.js";

// Early in request processing
if (rtkEnabled) {
  const rtkStats = compressMessages(body, true);
  if (rtkStats) {
    log.rtk = formatRtkLog(rtkStats);
  }
}
```

### Dashboard Settings

**Location**: `src/app/api/settings/route.js`

```javascript
{
  endpoint: {
    rtkEnabled: true,           // Toggle RTK
    cavemanEnabled: false,      // Toggle Caveman
    cavemanLevel: "medium",     // "low" | "medium" | "high"
  }
}
```

### Request Logging

**Location**: `open-sse/utils/requestLogger.js`

```javascript
{
  timestamp: "2026-05-11T10:30:00Z",
  model: "cc/claude-opus-4-5",
  rtk: {
    enabled: true,
    bytesBefore: 47000,
    bytesAfter: 28000,
    savings: "40%",
    filters: ["git-diff", "grep"],
  },
  tokens: {
    input: 1250,
    output: 850,
  }
}
```

---

## Performance Considerations

### Compression Overhead

- **Detection**: ~1-2ms (pattern matching on 1KB)
- **Filtering**: ~5-20ms (depends on output size)
- **Total**: <50ms for typical requests

### Memory Usage

- **In-Memory**: Only processes one message at a time
- **Streaming**: Doesn't buffer full response
- **Peak**: ~1MB for 100KB tool output

### Optimization Tips

1. **Disable for small outputs**: `MIN_COMPRESS_SIZE = 500`
2. **Cap raw size**: `RAW_CAP = 100_000`
3. **Skip Caveman for critical tasks**: Use `cavemanEnabled = false`
4. **Monitor compression ratio**: Track `bytesAfter / bytesBefore`

---

## Known Limitations

### RTK Limitations

1. **Markdown Tables**: May misalign column formatting
2. **ASCII Art**: Can break visual alignment
3. **Code Blocks**: May remove important context
4. **Custom Formats**: Only supports 10 built-in filters

### Caveman Limitations

1. **Quality Loss**: Reduces output quality for strict prompts
2. **Semantic Loss**: May lose nuance in error messages
3. **Non-English**: Not optimized for non-English output
4. **Structured Data**: May break JSON/YAML formatting

### Workarounds

- **Disable RTK per endpoint**: Dashboard → Endpoint Settings
- **Disable Caveman**: Keep `cavemanEnabled = false`
- **Custom filters**: Add new filter to `open-sse/rtk/filters/`
- **Bypass RTK**: Use `rtkEnabled = false` in request

---

## Testing

### Golden File Tests

**Location**: `tests/rtk/`

```
tests/rtk/
├── git-diff/
│   ├── in.txt
│   └── out.txt
├── grep/
│   ├── in.txt
│   └── out.txt
└── ...
```

### Test Execution

```bash
# Run RTK tests
npm test -- rtk

# Update golden files
npm test -- rtk --update
```

### Compression Ratio Tracking

```javascript
// In usage tracking
{
  rtk: {
    enabled: true,
    bytesBefore: 47000,
    bytesAfter: 28000,
    ratio: 0.60,  // 60% of original
    savings: 0.40, // 40% saved
  }
}
```

---

## References

- **RTK Source**: `open-sse/rtk/index.js`
- **Filters**: `open-sse/rtk/filters/`
- **Caveman**: `open-sse/rtk/caveman.js`
- **Integration**: `open-sse/handlers/chatCore.js`
- **Dashboard**: `src/app/api/settings/route.js`

---

**Document Version**: 1.0  
**9Router Version**: v0.4.29  
**Last Updated**: May 2026
