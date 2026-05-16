---
name: aiproxy-stt
description: Speech-to-text via AIPROXY — transcribe audio files. Use when user wants STT.
---

# AIPROXY STT

## Endpoint

```
POST /v1/audio/transcriptions
```

## Usage

```bash
curl -X POST $AIPROXY_URL/v1/audio/transcriptions \
  -H "Authorization: Bearer $AIPROXY_API_KEY" \
  -F "file=@audio.mp3" \
  -F "model=whisper-1"
```
