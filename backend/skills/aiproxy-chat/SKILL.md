---
name: aiproxy-chat
description: Chat completions via AIPROXY — streaming + non-streaming, multi-provider routing, tool calling, DCP, RTK compression. Use when user wants AI chat, code generation, or LLM completion.
---

# AIPROXY Chat

OpenAI-compatible chat completions with routing to 55+ providers.

## Endpoint

```
POST /v1/chat/completions
```

## Basic chat

```bash
curl -X POST $AIPROXY_URL/v1/chat/completions \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "kiro/claude-sonnet-4",
    "messages": [{"role": "user", "content": "hello"}],
    "stream": false
  }'
```

## Streaming

```bash
curl -X POST $AIPROXY_URL/v1/chat/completions \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-web/gemini-3-flash",
    "messages": [{"role": "user", "content": "count 1 to 5"}],
    "stream": true,
    "max_tokens": 100
  }'
```

## Anthropic format

```
POST /v1/messages
```

Same model format, supports Claude-native streaming.

## DCP (auto cleanup)

DCP runs automatically on every request — dedup tool calls, prune errors. Toggle:
```sql
INSERT INTO settings (key, value) VALUES ('dcpEnabled', 'false') ON CONFLICT (key) DO UPDATE SET value = 'false';
```

## Recommended providers

| Need | Provider |
|---|---|
| Fast/cheap | `opencode/deepseek-v4-flash-free` |
| Reasoning | `kiro/claude-sonnet-4` |
| Coding | `kiro/claude-sonnet-4` or `kiro/auto` |
| Free Gemini | `gemini-web/gemini-3-flash` |
