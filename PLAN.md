# AI-NPC Platform — Revised Plan

> **Foundation**: AIPROXY (AI Gateway + 9Router built-in)  
> **New**: Hermes (Queue/Worker) + Bot Layer + MCP Tools  
> **VPS**: 1 core / 1GB RAM / 40GB storage (Alibaba 24/7)

---

## Architecture

```
User
  ↓
Telegram / WhatsApp Bot
  ↓
Hermes (Asynq + Redis)
  ├── text_job → sync → AIPROXY → LLM → respond
  ├── image_job → async → AIPROXY → MCP image_api → save → notify
  ├── video_job → async → AIPROXY → MCP video_api → save → notify
  └── memory_job → async → update conversation history
         ↓
AIPROXY (AI Gateway)
  ├── 9Router (55+ providers)
  ├── Account pool + state machine
  ├── Format translation
  ├── Auth (JWT, API key, OAuth)
  ├── RTK compression + Caveman mode
  ├── Context compact
  └── REST API (OpenAI-compatible)
         ↓
External AI Providers (55+)
```

---

## Flow Detail

### Sync Flow (Chat)

```
User → Bot → Hermes (text_job)
  → AIPROXY /v1/chat/completions
  → LLM responds
  → Hermes sends back to Bot
```

### Async Flow (Image/Video)

```
User → Bot → Hermes (image_job)
  → Queue (Redis)
  → Worker picks up
  → AIPROXY (LLM decides tool)
  → MCP image_api → call external provider
  → Save result → notify user
```

---

## VPS Role

### Alibaba 1GB — runs 24/7:
```
AIPROXY             ~30 MB
PostgreSQL          ~60 MB
Redis + Asynq       ~50 MB
Bot (Telegram)      ~20 MB
Bot (WhatsApp)      ~20 MB
────────────────────────
Total               ~180 MB
```

### NOT on VPS (laptop/RDP only):
- Browser automation (Camoufox, Playwright)
- Add/refresh provider accounts
- Heavy build/compile (opsional)

---

## Components

### 1. AIPROXY (Existing — reuse as-is)
Already provides: chat completions, image gen, TTS/STT, embeddings, search, 55+ providers, account pool, streaming, auth, proxy pool, admin API, auto-sync 30s.

### 2. Hermes (New — Queue + Worker)
Asynq + Redis. Job types: text_job (sync), image_job (async), video_job (async), memory_job (async).

### 3. Bot Layer (New)
- Telegram: gotgbot/telegraf, webhook-based, stateless
- WhatsApp: whatsmeow, WebSocket, session persistence

### 4. MCP Servers (New)
image_api_mcp, video_api_mcp, telegram_mcp, whatsapp_mcp, memory_mcp, http_tools_mcp

### 5. Memory/NPC (Future)
Chat history, personality layers, user prefs. Prompt injection + PostgreSQL.

---

## Build Order

### Phase 0 — Foundation (DONE)
- [x] AIPROXY with 55+ providers, pool, streaming, image, auth, proxy
- [x] Auto-sync accounts every 30s

### Phase 1 — Queue + Bot (NEXT)
- [ ] Add Redis + Asynq deps
- [ ] Hermes package (queue types, enqueue, worker base)
- [ ] Telegram bot
- [ ] WhatsApp bot
- [ ] Sync text + Async image flows

### Phase 2 — MCP Tools
- [ ] MCP server base (JSON-RPC)
- [ ] image_api_mcp, http_tools_mcp, memory_mcp

### Phase 3 — Memory + NPC
- [ ] Conversation history, personality injection, user prefs

### Phase 4 — Polish
- [ ] Video generation, queue monitoring, graceful degradation

---

## Key Principles
1. AIPROXY is the core — semua AI routing via AIPROXY API
2. Sync for chat, Async for media
3. No heavy compute locally
4. Browser automation off-server (laptop/RDP)
5. Hermes is the brain, MCP is the hands
6. 1GB VPS enough (~180MB baseline)

---

## Open Questions
- [ ] Hermes: binary terpisah atau bagian dari AIPROXY?
- [ ] MCP: stdio atau HTTP?
- [ ] WhatsApp session: PostgreSQL atau file?
- [ ] Queue fallback: Redis down → SQLite queue?
- [ ] Bot webhook: public endpoint atau polling?
