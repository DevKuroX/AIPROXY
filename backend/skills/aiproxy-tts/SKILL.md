---
name: aiproxy-tts
description: Text-to-speech via AIPROXY — convert text to audio. Use when user wants TTS.
---

# AIPROXY TTS

## Endpoint

```
POST /v1/audio/speech
```

## Usage

```bash
curl -X POST $AIPROXY_URL/v1/audio/speech \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"model":"tts-1","input":"Hello world","voice":"alloy"}'
```
