---
name: aiproxy-image
description: Image generation via AIPROXY — 12 providers (StabilityAI, FalAI, OpenAI DALL-E, Gemini, RunwayML). Use when user wants image generation.
---

# AIPROXY Image

OpenAI-compatible image generation with 12 provider adapters.

## Endpoint

```
POST /v1/images/generations
```

## Generate

```bash
curl -X POST $AIPROXY_URL/v1/images/generations \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"dall-e-3","prompt":"a cat","n":1,"size":"1024x1024"}'
```

## Providers

| Model | Provider |
|---|---|
| `dall-e-3` | OpenAI |
| `stabilityai/stable-diffusion-xl` | StabilityAI |
| `fal-ai/flux` | FalAI |
| `gemini-3-pro` | Gemini |
| `black-forest-labs/flux` | Black Forest Labs |
