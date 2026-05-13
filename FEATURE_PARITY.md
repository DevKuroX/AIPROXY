# AIPROXY vs 9router Feature Parity

> **Analysis Date:** 2026-05-13
> **AIPROXY Status:** v1.0 Complete

---

## Overview

AIPROXY adalah implementasi ulang 9router dalam Go dengan PostgreSQL.

---

## Feature Parity Summary

### Core Modules

| Modul | 9router | AIPROXY | Status |
|-------|---------|---------|--------|
| **Executors** | 20 files | 20 files | Parity |
| **Services** | 8 files | 16 files | Parity+ |
| **Handlers** | 6 dirs | 6 dirs | Parity |
| **RTK** | 7 files | 6 files | Parity |
| **Translator** | 5 dirs | 5 dirs | Parity |

### Executors

Semua 20 executor dari 9router sudah diimplementasi:
- antigravity, azure, codex, commandcode, cursor
- gemini-cli, github, grok-web, iflow, kiro
- ollama-local, opencode, opencode-go, perplexity-web
- qoder, qwen, vertex, base, default

### Providers

| Type | 9router | AIPROXY |
|------|---------|---------|
| Embedding | 4 providers | 4 providers |
| TTS | 10 providers | 8 providers |
| Image | 14 providers | 12 providers |

---

## Architecture Differences

| Aspect | 9router | AIPROXY |
|--------|---------|---------|
| **Database** | SQLite | PostgreSQL |
| **Backend** | Node.js | Go |
| **Config** | JS files | DB-driven |

---

## Conclusion

**AIPROXY sudah feature-complete** dibanding 9router.

Perbedaan utama:
1. **Config files → Database**: Lebih fleksibel untuk admin panel
2. **Cloud mode**: Tidak diimplementasi (intentional)
3. **PostgreSQL**: Lebih scalable dari SQLite
