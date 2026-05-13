// Package rtk contains runtime kit utilities for request transformation.
// ref: open-sse/rtk/cavemanPrompts.js
package rtk

// Caveman intensity levels for prompt injection.
// ref: open-sse/rtk/cavemanPrompts.js:4-8
const (
	CavemanLevelLite  = "lite"
	CavemanLevelFull  = "full"
	CavemanLevelUltra = "ultra"
)

// Shared boundaries text - exact content that must remain unchanged.
// ref: open-sse/rtk/cavemanPrompts.js:10
const sharedBoundaries = "Code blocks, file paths, commands, errors, URLs: keep exact. Security warnings, irreversible action confirmations, multi-step ordered sequences: write normal. Resume terse style after."

// cavemanPrompts maps level names to their respective prompt strings.
// ref: open-sse/rtk/cavemanPrompts.js:12-34
var cavemanPrompts = map[string]string{
	// ref: open-sse/rtk/cavemanPrompts.js:13-18
	CavemanLevelLite: joinPrompts(
		"Respond tersely. Keep grammar and full sentences but drop filler, hedging and pleasantries (just/really/basically/sure/of course/I'd be happy to).",
		"Pattern: state the thing, the action, the reason. Then next step.",
		sharedBoundaries,
		"Active every response until user asks for normal mode.",
	),

	// ref: open-sse/rtk/cavemanPrompts.js:20-26
	CavemanLevelFull: joinPrompts(
		"Respond like terse caveman. All technical substance stay exact, only fluff die.",
		"Drop: articles (a/an/the), filler (just/really/basically/actually/simply), pleasantries, hedging. Fragments OK. Short synonyms (big not extensive, fix not implement a solution for).",
		"Pattern: [thing] [action] [reason]. [next step].",
		sharedBoundaries,
		"Active every response until user asks for normal mode.",
	),

	// ref: open-sse/rtk/cavemanPrompts.js:28-34
	CavemanLevelUltra: joinPrompts(
		"Respond ultra-terse. Maximum compression. Telegraphic.",
		"Abbreviate (DB/auth/config/req/res/fn/impl), strip conjunctions, use arrows for causality (X → Y). One word when one word enough.",
		"Pattern: [thing] → [result]. [fix].",
		sharedBoundaries,
		"Active every response until user asks for normal mode.",
	),
}

// joinPrompts concatenates prompt parts with space separator.
// ref: open-sse/rtk/cavemanPrompts.js:18,26,34 (join pattern)
func joinPrompts(parts ...string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += " "
		}
		result += part
	}
	return result
}

// GetCavemanPrompt returns the caveman prompt for the given level.
// Returns empty string if level is invalid.
func GetCavemanPrompt(level string) string {
	if prompt, ok := cavemanPrompts[level]; ok {
		return prompt
	}
	return ""
}
