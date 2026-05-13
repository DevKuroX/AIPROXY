# docs/CAVEMAN_SPEC.md — Caveman Prompt Injector (Parity Spec)

> Port of `_ref/9router/open-sse/rtk/caveman.js` and `cavemanPrompts.js`.

---

## 1. What Caveman does

Caveman is a **system-prompt injector** that nudges the model to produce
terse output, saving **completion tokens** (where RTK saves **prompt** tokens).

It appends a brevity instruction to the system / developer message of every
outgoing request. Three intensity levels: `LITE`, `FULL`, `ULTRA`.

The instruction is dispatched by **output format** so it works for both
translated and native-passthrough flows.

Reference: adapted from <https://github.com/JuliusBrussee/caveman>.

---

## 2. When Caveman runs

```
incoming body
   │
   ▼
RTK.CompressMessages    ← prompt-token saver
   │
   ▼
translator.Convert       ← shape conversion
   │
   ▼
Caveman.Inject           ← THIS step — runs on final-shape body
   │
   ▼
executor.Send
```

Caveman runs **after** translation, on the body that is about to hit the wire.
This is important because the system-prompt location differs per format
(OpenAI uses `messages[]`, Claude uses `system`, Gemini uses
`systemInstruction`, OpenAI Responses uses `instructions` string).

---

## 3. Toggling

| Source | Effect |
|---|---|
| `settings.caveman_level` (`off`, `lite`, `full`, `ultra`) | Global default |
| Request header `X-9R-Caveman: off` | Disable for this request |
| Request header `X-9R-Caveman: lite|full|ultra` | Override level |

If level is `off`, Inject is a no-op.

---

## 4. The prompts (VERBATIM — do not paraphrase)

Port `cavemanPrompts.js` exactly. These prompts are tuned — any rewording
changes behavior. Embed them as raw string constants.

```go
// internal/caveman/prompts.go
package caveman

type Level string

const (
    LevelOff   Level = "off"
    LevelLite  Level = "lite"
    LevelFull  Level = "full"
    LevelUltra Level = "ultra"
)

const sharedBoundaries = "Code blocks, file paths, commands, errors, URLs: keep exact. Security warnings, irreversible action confirmations, multi-step ordered sequences: write normal. Resume terse style after."

var prompts = map[Level]string{
    LevelLite: strings.Join([]string{
        "Respond tersely. Keep grammar and full sentences but drop filler, hedging and pleasantries (just/really/basically/sure/of course/I'd be happy to).",
        "Pattern: state the thing, the action, the reason. Then next step.",
        sharedBoundaries,
        "Active every response until user asks for normal mode.",
    }, " "),

    LevelFull: strings.Join([]string{
        "Respond like terse caveman. All technical substance stay exact, only fluff die.",
        "Drop: articles (a/an/the), filler (just/really/basically/actually/simply), pleasantries, hedging. Fragments OK. Short synonyms (big not extensive, fix not implement a solution for).",
        "Pattern: [thing] [action] [reason]. [next step].",
        sharedBoundaries,
        "Active every response until user asks for normal mode.",
    }, " "),

    LevelUltra: strings.Join([]string{
        "Respond ultra-terse. Maximum compression. Telegraphic.",
        "Abbreviate (DB/auth/config/req/res/fn/impl), strip conjunctions, use arrows for causality (X → Y). One word when one word enough.",
        "Pattern: [thing] → [result]. [fix].",
        sharedBoundaries,
        "Active every response until user asks for normal mode.",
    }, " "),
}

func PromptFor(l Level) (string, bool) {
    s, ok := prompts[l]
    return s, ok
}
```

---

## 5. Per-format injection logic

Port `caveman.js`'s switch statement. Separator is **always** `"\n\n"`.

```go
// internal/caveman/inject.go
package caveman

import (
    "9rgo/internal/translator"
)

const sep = "\n\n"

func Inject(body map[string]any, format translator.Format, level Level) {
    prompt, ok := PromptFor(level)
    if !ok || body == nil { return }

    switch format {
    case translator.FormatClaude:
        injectClaudeSystem(body, prompt)
    case translator.FormatGemini,
         translator.FormatGeminiCLI,
         translator.FormatVertex,
         translator.FormatAntigravity:
        injectGeminiSystem(body, prompt)
    default:
        // OpenAI and OpenAI-shaped formats:
        // responses, codex, cursor, kiro, ollama, commandcode
        injectMessagesSystem(body, prompt)
    }
}
```

### 5.1 OpenAI-shaped (`injectMessagesSystem`)

Handles three sub-cases (mirror `caveman.js` `injectMessagesSystem`):

1. **OpenAI Responses with `instructions` string field**
   ```text
   if body.instructions is string:
       body.instructions = (existing + SEP + prompt) if non-empty else prompt
       return
   ```

2. **Has `messages[]` or `input[]` array**:
   - Find first index where `role == "system"` OR `role == "developer"`.
   - If found → append (`appendToOpenAIMessage`).
   - Else → prepend new `{role: "system", content: prompt}`.

3. **Neither container** → no-op.

#### `appendToOpenAIMessage`
- If `content` is string → `content = old + SEP + prompt`.
- If `content` is array → push `{type: "text", text: prompt}`.

### 5.2 Claude (`injectClaudeSystem`)

Claude has a top-level `system` field that is either a string or an array
of `{type: "text", text: "..."}` blocks.

- If `system` is undefined → set `system = prompt` (string).
- If `system` is string → `system = old + SEP + prompt`.
- If `system` is array → append `{type: "text", text: prompt}`.

### 5.3 Gemini family (`injectGeminiSystem`)

Gemini uses `systemInstruction: { role: "system", parts: [{text: "..."}] }`.

**Special handling for Antigravity**: the body wraps Gemini under
`body.request`. If `body.request` exists, operate on it.

- If `systemInstruction` missing → create `{ role: "system", parts: [{text: prompt}] }`.
- If `parts` exists → append `{text: prompt}`.
- If `parts` missing but `systemInstruction` exists → set `parts = [{text: prompt}]`.

---

## 6. Edge cases (port carefully)

| Case | Behavior |
|---|---|
| Empty body | No-op |
| Body has no system anywhere | Prepend (OpenAI) / set (Claude/Gemini) |
| System already very long | Append regardless; provider will truncate if needed |
| Multiple system messages (OpenAI sometimes) | Append to the FIRST one only |
| `role: "developer"` (OpenAI new) | Treat as system |
| OpenAI Responses with both `instructions` AND `input[]` | Prefer `instructions` (per 9router) |
| Antigravity nested body | Operate on `body.request` |

---

## 7. Public API

```go
package caveman

type Level string

const (
    LevelOff   Level = "off"
    LevelLite  Level = "lite"
    LevelFull  Level = "full"
    LevelUltra Level = "ultra"
)

// PromptFor returns the prompt string for a level, or ("", false) if off/invalid.
func PromptFor(l Level) (string, bool)

// Inject mutates body in place, adding the caveman prompt for the given format.
// If level == LevelOff or body is nil, it is a no-op.
func Inject(body map[string]any, format translator.Format, level Level)

// ParseLevel converts a string to a Level (defensive parsing).
func ParseLevel(s string) Level
```

---

## 8. Testing

### Golden fixtures

Layout:
```
testdata/caveman/<level>/<format>/
├── in.json     # body before injection
└── out.json    # body after injection
```

Levels: `lite`, `full`, `ultra` (skip `off` — no-op).
Formats: `openai-chat`, `openai-responses-string`, `openai-responses-array`,
`claude-string`, `claude-array`, `gemini`, `antigravity`.

→ Total: 3 × 7 = 21 baseline cases.

Plus edge cases:
- No system message
- Multiple system messages (OpenAI)
- Developer role
- Empty content
- Pre-existing very long system

### Test runner

```go
func TestInject(t *testing.T) {
    levels := []caveman.Level{caveman.LevelLite, caveman.LevelFull, caveman.LevelUltra}
    formats := map[string]translator.Format{
        "openai-chat":   translator.FormatOpenAI,
        "openai-resp":   translator.FormatOpenAIResponses,
        "claude":        translator.FormatClaude,
        "gemini":        translator.FormatGemini,
        "antigravity":   translator.FormatAntigravity,
    }
    for _, lvl := range levels {
        for name, fmt := range formats {
            t.Run(fmt.Sprintf("%s/%s", lvl, name), func(t *testing.T) {
                in   := loadJSON(t, "testdata/caveman", string(lvl), name, "in.json")
                want := loadJSON(t, "testdata/caveman", string(lvl), name, "out.json")
                caveman.Inject(in, fmt, lvl)
                require.Equal(t, want, in)
            })
        }
    }
}
```

---

## 9. RTK + Caveman interaction

Both can be on simultaneously. Order is:
1. RTK compresses tool-results in the **source-format** body.
2. Translator converts shape.
3. Caveman injects in the **target-format** body.

This means RTK only sees `tool_result` blocks (it doesn't care about system
messages), and Caveman only sees system messages (it doesn't care about tool
content). They never conflict.

---

## 10. Reference

| Concept | 9router file |
|---|---|
| Injector | `open-sse/rtk/caveman.js` |
| Prompts | `open-sse/rtk/cavemanPrompts.js` |
| Format constants | `open-sse/translator/formats.js` |
| Pipeline integration | `open-sse/handlers/chatCore.js` |
