# 9Router Research Documentation Index

## Overview

This index provides a comprehensive guide to the 9Router project research documentation. All documents are located in `/home/ubuntu/`.

---

## Documents

### 1. **9router_research.md** (740 lines)
**Comprehensive Architecture & Structure Guide**

Complete breakdown of the 9Router project including:
- Executive summary
- Directory structure (detailed)
- Core modules (open-sse, src/app, src/lib, src/mitm)
- Key features & systems
- Configuration files
- API endpoints
- Utilities & helpers
- Testing framework
- Build & deployment
- Dependencies
- Key concepts
- Recent updates
- Key takeaways for phase prompts

**Use this for**: Deep understanding of the entire architecture

---

### 2. **9router_architecture_summary.txt** (ASCII Art)
**Visual Architecture Overview**

High-level visual representation including:
- System overview diagram
- Core modules breakdown
- Key features summary
- Provider support matrix
- API endpoints list
- Database schema
- Technology stack
- File statistics
- Development commands

**Use this for**: Quick visual reference and presentations

---

### 3. **9router_phase_prompt_guidelines.md** (500+ lines)
**Guidelines for Creating Accurate Phase Prompts**

Detailed guidelines for creating phase prompts including:
- Key architectural principles (7 principles)
- Phase prompt categories (10 categories A-J)
- Prompt template
- Common pitfalls to avoid (8 pitfalls)
- Phase prompt examples (2 detailed examples)
- Checklist for phase prompt creation

**Categories covered**:
- A: Core Routing Logic
- B: Provider Integration
- C: Format Translation
- D: Token Optimization (RTK)
- E: Database Operations
- F: API Endpoints
- G: Dashboard UI
- H: MITM Proxy System
- I: OAuth & Token Management
- J: Media Providers

**Use this for**: Creating accurate, contextually appropriate phase prompts

---

### 4. **9router_quick_reference.md** (300+ lines)
**Quick Reference Guide**

Condensed reference guide including:
- Project overview
- Architecture at a glance
- Core modules table
- Key features summary
- Provider support list
- API endpoints reference
- File organization
- Database schema
- Technology stack
- Development commands
- Key concepts
- Common tasks
- Statistics
- Important files
- Useful links
- Notes for phase prompts

**Use this for**: Quick lookups and common tasks

---

## How to Use These Documents

### For Understanding the Project
1. Start with **9router_quick_reference.md** for overview
2. Read **9router_architecture_summary.txt** for visual understanding
3. Deep dive into **9router_research.md** for details

### For Creating Phase Prompts
1. Review **9router_phase_prompt_guidelines.md** for principles
2. Check **9router_quick_reference.md** for file locations
3. Reference **9router_research.md** for implementation details
4. Use the prompt template from guidelines

### For Specific Tasks
- **Adding a provider**: See "Category B" in guidelines + "Add a New Provider" in quick reference
- **Adding RTK filter**: See "Category D" in guidelines + "Add RTK Compression Filter" in quick reference
- **Adding API endpoint**: See "Category F" in guidelines + "Add API Endpoint" in quick reference
- **Adding dashboard page**: See "Category G" in guidelines + "Add Dashboard Page" in quick reference
- **Adding database repo**: See "Category E" in guidelines + "Add Database Repository" in quick reference

---

## Key Architectural Principles (Summary)

1. **Modular Separation**: open-sse/ (provider-agnostic) vs src/ (Next.js-specific)
2. **Provider Abstraction**: Executor + Format + Translator + OAuth Service
3. **Format Translation**: 9 formats with bidirectional translation
4. **Multi-Tier Fallback**: Subscription → Cheap → Free
5. **Token Optimization (RTK)**: 20-40% savings via compression
6. **Database Abstraction**: 4 adapters for portability
7. **MITM Proxy System**: Optional transparent interception

---

## Core Modules (Quick Reference)

| Module | Purpose | Files |
|--------|---------|-------|
| **open-sse/** | Provider-agnostic routing | config/, executors/, handlers/, services/, translator/, rtk/, utils/ |
| **src/app/api/** | 0penAI-compatible API | v1/, v1beta/, management APIs |
| **src/app/(dashboard)/** | React dashboard | providers/, combos/, cli-tools/, mitm/ |
| **src/lib/db/** | Database layer | driver.js, repos/, adapters/, migrations/ |
| **src/lib/oauth/** | OAuth (13 providers) | services/, utils/ |
| **src/mitm/** | MITM proxy | server.js, handlers/, cert/, dns/ |

---

## Provider Support (Quick Reference)

- **OAuth Providers**: 13 (CL4ude, Gemini, Codex, Cursor, GitHub, Antigravity, Kiro, iFlow, Qwen, OpenCode, etc.)
- **API Key Providers**: 6 (0penAI, OpenRouter, GLM, Kimi, MiniMax, Azure)
- **Compatible Nodes**: Ollama, 0penAI-compatible endpoints
- **Web Scrapers**: Grok, Perplexity
- **Total**: 40+ providers

---

## Format Support (Quick Reference)

9 formats with bidirectional translation:
1. 0penAI (standard)
2. CL4ude ()
3. Gemini (Google)
4. Responses API (0penAI)
5. Cursor
6. Antigravity
7. Kiro
8. CommandCode
9. Ollama

---

## Media Providers (Quick Reference)

- **Image Generation**: 14 providers
- **Text-to-Speech**: 10 providers
- **Embeddings**: 5 providers
- **Speech-to-Text**: STT support

---

## Database Repositories (Quick Reference)

10 data types:
1. connections - Provider auth
2. apiKeys - Stored keys
3. combos - Fallback sequences
4. aliases - Model aliases
5. settings - User settings
6. pricing - Model pricing
7. requestDetails - Request logs
8. disabledModels - Disabled models
9. proxyPools - Proxy pools
10. nodes - Compatible nodes

---

## Database Adapters (Quick Reference)

4 adapters for different environments:
1. **better-sqlite3** - Primary (fastest, requires build tools)
2. **sql.js** - Fallback (pure JS, no build required)
3. **Node.js sqlite** - Alternative
4. **Bun sqlite** - Bun runtime support

---

## RTK Compression Filters (Quick Reference)

10 filters for token optimization:
1. grep - Grep output compression
2. ls - Directory listing compression
3. gitDiff - Git diff compression
4. gitStatus - Git status compression
5. tree - Tree output compression
6. readNumbered - Numbered content compression
7. searchList - Search result compression
8. smartTruncate - Smart truncation
9. find - Find command compression
10. dedupLog - Deduplication

---

## Important Files (Quick Reference)

### Core Routing
- `open-sse/handlers/chatCore.js` - Main chat handler
- `open-sse/services/provider.js` - Provider config lookup
- `open-sse/services/accountFallback.js` - Fallback logic
- `open-sse/translator/index.js` - Format translation

### Providers
- `open-sse/executors/index.js` - Executor registry
- `open-sse/config/providers.js` - Provider definitions
- `open-sse/config/providerModels.js` - Model mapping

### Token Optimization
- `open-sse/rtk/index.js` - RTK orchestration
- `open-sse/rtk/caveman.js` - Compression algorithm
- `open-sse/rtk/registry.js` - Filter registry

### Database
- `src/lib/db/driver.js` - Database driver
- `src/lib/db/repos/` - Data repositories
- `src/lib/db/adapters/` - Database adapters

### OAuth
- `src/lib/oauth/services/` - OAuth implementations
- `open-sse/services/tokenRefresh.js` - Token refresh

### MITM
- `src/mitm/server.js` - MITM proxy server
- `src/mitm/handlers/` - Provider-specific handlers
- `src/mitm/cert/` - Certificate management

### API
- `src/app/api/v1/` - 0penAI-compatible endpoints
- `src/app/api/providers/` - Provider management
- `src/app/api/oauth/` - OAuth flow

### Dashboard
- `src/app/(dashboard)/dashboard/` - Dashboard pages
- `src/store/` - Zustand state

---

## Statistics

- **Total LOC**: ~17,767
- **JavaScript Files**: 82+
- **Providers**: 40+
- **Image Providers**: 14
- **TTS Providers**: 10
- **Embedding Providers**: 5
- **Test Files**: 25+
- **Database Adapters**: 4
- **OAuth Providers**: 13
- **Format Translators**: 9
- **RTK Filters**: 10
- **Contributors**: 90+

---

## Technology Stack

### Frontend
- Next.js 16.1.6
- React 19.2.4
- Tailwind CSS 4
- Monaco Editor
- Recharts
- Zustand

### Backend
- Next.js API routes
- Express 5.2.1
- Node.js / Bun
- SQLite (4 adapters)
- node-forge
- http-proxy-middleware
- jose
- bcryptjs
- socks-proxy-agent

---

## Common Tasks

### Add a New Provider
1. Create executor: `open-sse/executors/[provider].js`
2. Add config: `open-sse/config/providers.js`
3. Implement OAuth: `src/lib/oauth/services/[provider].js`
4. Add translators: `open-sse/translator/request/` & `response/`
5. Add tests

### Add RTK Compression Filter
1. Create filter: `open-sse/rtk/filters/[name].js`
2. Register: `open-sse/rtk/registry.js`
3. Add tests

### Add API Endpoint
1. Create route: `src/app/api/[path]/route.js`
2. Use database repos
3. Handle errors
4. Add tests

### Add Dashboard Page
1. Create page: `src/app/(dashboard)/dashboard/[page]/page.js`
2. Use React components
3. Call management APIs
4. Use Zustand for state

### Add Database Repository
1. Create repo: `src/lib/db/repos/[entity]Repo.js`
2. Implement CRUD
3. Support all 4 adapters
4. Add migration if needed

---

## Notes for Phase Prompts

1. **Always consider all 40+ providers** when implementing features
2. **Always consider all 9 formats** for translation
3. **Always respect fallback tiers** (subscription → cheap → free)
4. **Always apply RTK compression** before sending requests
5. **Always support all 4 database adapters**
6. **Always handle errors gracefully** with fallback
7. **Always test with multiple providers** (OAuth, API key, compatible node)
8. **Always support streaming** (SSE) and non-streaming (JSON)
9. **Always implement OAuth token refresh** for OAuth providers
10. **Always document API endpoints** with request/response format

---

## Useful Links

- **GitHub**: https://github.com/decolua/9router
- **Website**: https://9router.com
- **Issues**: https://github.com/decolua/9router/issues
- **Architecture Docs**: https://github.com/decolua/9router/blob/master/docs/ARCHITECTURE.md

---

## Document Locations

All research documents are saved in `/home/ubuntu/`:

```
/home/ubuntu/
├── 9router_research.md                    # Comprehensive guide (740 lines)
├── 9router_architecture_summary.txt       # Visual overview (ASCII art)
├── 9router_phase_prompt_guidelines.md     # Phase prompt guidelines (500+ lines)
├── 9router_quick_reference.md             # Quick reference (300+ lines)
└── 9router_research_INDEX.md              # This file
```

---

## How to Get Started

1. **First time?** Read `9router_quick_reference.md` (5 min)
2. **Need visual?** Check `9router_architecture_summary.txt` (5 min)
3. **Creating prompts?** Study `9router_phase_prompt_guidelines.md` (15 min)
4. **Deep dive?** Read `9router_research.md` (30 min)

---

## Questions?

Refer to the appropriate document:
- **"What is 9Router?"** → quick_reference.md
- **"How is it structured?"** → architecture_summary.txt
- **"How do I create a phase prompt?"** → phase_prompt_guidelines.md
- **"What are the details?"** → research.md

---

**Last Updated**: 2026-05-12  
**Research Completed**: May 12, 2026  
**Repository**: decolua/9router v0.4.31

