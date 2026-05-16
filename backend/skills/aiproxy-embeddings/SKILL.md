---
name: aiproxy-embeddings
description: Text embeddings via AIPROXY. Use when user wants vector embeddings.
---

# AIPROXY Embeddings

## Endpoint

```
POST /v1/embeddings
```

## Usage

```bash
curl -X POST $AIPROXY_URL/v1/embeddings \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"text-embedding-3-small","input":"hello world"}'
```
