---
name: aiproxy-web
description: Web search + fetch via AIPROXY. Use when user wants web search or URL fetching.
---

# AIPROXY Web

## Search

```
POST /v1/search
```

```bash
curl -X POST $AIPROXY_URL/v1/search \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query":"latest AI news"}'
```

## Fetch (URL → markdown)

```
POST /v1/fetch
```

```bash
curl -X POST $AIPROXY_URL/v1/fetch \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com"}'
```
