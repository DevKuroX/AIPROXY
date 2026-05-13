# 9Router Documentation Resources

**Compiled**: May 2026  
**9Router Version**: v0.4.29  
**Status**: Complete Reference Package

---

## 📚 Documentation Files Created

### 1. **9ROUTER_REFERENCE.md** (Comprehensive Overview)
- Executive summary of 9Router
- Architecture overview with diagrams
- All key features explained (RTK, Fallback, Combos, Quota Tracking, etc.)
- Data model & storage schema
- API endpoints (compatibility + management)
- Request/response flow diagrams
- Configuration & environment variables
- Deployment options
- Known quirks & limitations
- Technology stack
- Community & support links

**Use for**: Understanding what 9Router does and how it works at a high level.

### 2. **9ROUTER_CODEBASE_MAP.md** (Code Structure)
- Complete directory structure
- File organization by purpose
- Core components breakdown
- Data flow diagrams
- Critical files for porting
- Provider executors reference
- RTK filters reference
- Configuration & constants
- Testing & fixtures
- Build & deployment
- Key patterns & conventions
- Known quirks for porting

**Use for**: Navigating the 9router codebase and understanding code organization.

### 3. **9ROUTER_RTK_CAVEMAN.md** (Token Compression Deep Dive)
- RTK architecture & data flow
- Core compression functions
- All 10 compression filters explained
- Caveman mode (aggressive compression)
- Configuration & constants
- Integration points
- Performance considerations
- Known limitations & workarounds
- Testing approach
- Real-world examples

**Use for**: Understanding token compression and implementing RTK in Go.

---

## 🔗 Official 9Router Resources

### GitHub Repository
- **Main Repo**: https://github.com/decolua/9router
- **Latest Release**: v0.4.29 (May 2026)
- **License**: MIT
- **Stars**: 6,874 | **Forks**: 1,203

### Official Documentation
- **README**: https://github.com/decolua/9router/blob/master/README.md
- **Architecture**: https://github.com/decolua/9router/blob/master/docs/ARCHITECTURE.md
- **Changelog**: https://github.com/decolua/9router/blob/master/CHANGELOG.md
- **Docker Guide**: https://github.com/decolua/9router/blob/master/DOCKER.md

### Feature Guides (GitBook)
- **Quota Tracking**: gitbook/content/en/features/quota-tracking.md
- **Smart Routing**: gitbook/content/en/features/smart-routing.md
- **Combos**: gitbook/content/en/features/combos.md

### Website
- **9router.com**: https://9router.com
- **NPM Package**: https://www.npmjs.com/package/9router

---

## 📂 Local Reference Copy

**Location**: `_ref/9router/`

### Key Files to Read
```
_ref/9router/
├── README.md                     # Main documentation
├── CHANGELOG.md                  # Version history
├── docs/ARCHITECTURE.md          # System design
├── open-sse/handlers/chatCore.js # Core orchestration
├── open-sse/rtk/index.js         # Token compression
├── open-sse/translator/index.js  # Format translation
├── open-sse/services/accountFallback.js  # Fallback logic
└── src/sse/handlers/chat.js      # Request entry point
```

### Provider Executors (Reference)
```
_ref/9router/open-sse/executors/
├── default.js                    # Generic 0penAI-compatible (template)
├── cursor.js                     # Complex executor (22.4K)
├── github.js                     # OAuth provider (14.3K)
├── antigravity.js                # Google Antigravity (17.7K)
└── [14 more providers]
```

### RTK Filters (Reference)
```
_ref/9router/open-sse/rtk/
├── index.js                      # Main compression engine
├── autodetect.js                 # Format detection
├── caveman.js                    # Aggressive compression
└── filters/
    ├── git-diff.js
    ├── grep.js
    ├── find.js
    ├── ls.js
    ├── tree.js
    └── [5 more filters]
```

---

## 🎯 Quick Navigation by Topic

### Understanding 9Router
1. Start: **9ROUTER_REFERENCE.md** → Executive Overview
2. Deep dive: **9ROUTER_REFERENCE.md** → Architecture Overview
3. Features: **9ROUTER_REFERENCE.md** → Key Features section

### Porting to Go
1. Architecture: **9ROUTER_CODEBASE_MAP.md** → Directory Structure
2. Core logic: **9ROUTER_CODEBASE_MAP.md** → Critical Files for Porting
3. Token compression: **9ROUTER_RTK_CAVEMAN.md** → Full RTK guide
4. Reference code: `_ref/9router/open-sse/handlers/chatCore.js`

### Implementing RTK
1. Overview: **9ROUTER_RTK_CAVEMAN.md** → RTK Overview
2. Filters: **9ROUTER_RTK_CAVEMAN.md** → Compression Filters
3. Integration: **9ROUTER_RTK_CAVEMAN.md** → Integration Points
4. Reference: `_ref/9router/open-sse/rtk/`

### Understanding Fallback
1. Concept: **9ROUTER_REFERENCE.md** → Smart 3-Tier Fallback Routing
2. Logic: **9ROUTER_CODEBASE_MAP.md** → Fallback Flow diagram
3. Code: `_ref/9router/open-sse/services/accountFallback.js`

### Format Translation
1. Overview: **9ROUTER_REFERENCE.md** → Format Translation
2. Architecture: **9ROUTER_CODEBASE_MAP.md** → Translator section
3. Code: `_ref/9router/open-sse/translator/index.js`

### Provider Executors
1. Overview: **9ROUTER_REFERENCE.md** → Provider Executors
2. Structure: **9ROUTER_CODEBASE_MAP.md** → Executors section
3. Examples: `_ref/9router/open-sse/executors/`

### API Endpoints
1. Reference: **9ROUTER_REFERENCE.md** → API Endpoints
2. Implementation: `_ref/9router/src/app/api/`

### Quota Tracking
1. Features: **9ROUTER_REFERENCE.md** → Real-Time Quota Tracking
2. Data model: **9ROUTER_REFERENCE.md** → Data Model & Storage
3. Code: `_ref/9router/src/lib/usageDb.js`

---

## 🔍 Key Concepts Explained

### RTK (Rust Token Killer)
- **What**: Lossless compression for tool outputs
- **Where**: `open-sse/rtk/`
- **Why**: Save 20-40% input tokens
- **How**: Auto-detect format → apply filter → measure → keep if smaller
- **Read**: **9ROUTER_RTK_CAVEMAN.md**

### 3-Tier Fallback
- **What**: Automatic provider switching (Subscription → Cheap → Free)
- **Where**: `open-sse/services/accountFallback.js`
- **Why**: Never stop coding, minimize costs
- **How**: Check quota → select next tier → retry
- **Read**: **9ROUTER_REFERENCE.md** → Smart 3-Tier Fallback Routing

### Format Translation
- **What**: Convert between provider API formats
- **Where**: `open-sse/translator/`
- **Why**: Support 40+ providers with single endpoint
- **How**: Detect source → translate to provider → execute → translate back
- **Read**: **9ROUTER_REFERENCE.md** → Format Translation

### Combos
- **What**: User-defined fallback chains
- **Where**: `src/app/api/combos/`
- **Why**: Customize routing strategy
- **How**: Define sequence of models → 9Router tries in order
- **Read**: **9ROUTER_REFERENCE.md** → Combos - Custom Fallback Chains

### Quota Tracking
- **What**: Real-time token consumption monitoring
- **Where**: `src/lib/usageDb.js`
- **Why**: Maximize subscription value
- **How**: Track tokens per request → aggregate by provider → show in dashboard
- **Read**: **9ROUTER_REFERENCE.md** → Real-Time Quota Tracking

---

## 📊 Architecture Diagrams

### System Context
```
CLI Tools → 9Router → [Tier 1: Subscription]
                      [Tier 2: Cheap]
                      [Tier 3: Free]
```
**Location**: **9ROUTER_REFERENCE.md** → Architecture Overview

### Request Processing Pipeline
```
Request → Parse → RTK Compress → Format Translate → Execute → Track Usage
```
**Location**: **9ROUTER_CODEBASE_MAP.md** → Request Processing Pipeline

### Fallback Flow
```
Request fails → Parse error → Select next provider → Retry
```
**Location**: **9ROUTER_CODEBASE_MAP.md** → Fallback Flow

### RTK Compression
```
Detect format → Apply filter → Measure → Keep if smaller → Continue
```
**Location**: **9ROUTER_RTK_CAVEMAN.md** → Data Flow

---

## 🛠️ Technology Stack

### Backend
- **Runtime**: Node.js
- **Framework**: Next.js 16.1.6
- **Database**: LowDB (JSON) + optional better-sqlite3
- **Auth**: JWT (jose), bcryptjs
- **Streaming**: Native Node.js streams

### Frontend
- **Framework**: React 19.2.4
- **Styling**: Tailwind CSS 4
- **Charts**: Recharts 3.7.0
- **State**: Zustand 5.0.10

**Read**: **9ROUTER_REFERENCE.md** → Technology Stack

---

## 📋 Porting Checklist

### Phase 1: Foundation
- [ ] Read **9ROUTER_REFERENCE.md** (Executive Overview)
- [ ] Read **9ROUTER_CODEBASE_MAP.md** (Directory Structure)
- [ ] Set up Go module & chi router
- [ ] Implement basic HTTP server on port 20128

### Phase 2: Core Routing
- [ ] Implement request parser (src/sse/handlers/chat.js)
- [ ] Implement format detection (open-sse/services/provider.js)
- [ ] Implement model resolution (open-sse/services/model.js)
- [ ] Implement executor dispatch (open-sse/executors/index.js)

### Phase 3: RTK Token Compression
- [ ] Read **9ROUTER_RTK_CAVEMAN.md** (Full guide)
- [ ] Implement auto-detection (autodetect.js)
- [ ] Implement 10 compression filters
- [ ] Implement caveman mode (optional)

### Phase 4: Format Translation
- [ ] Implement translator registry (open-sse/translator/index.js)
- [ ] Implement request translators (5+ formats)
- [ ] Implement response translators (5+ formats)
- [ ] Test with real provider APIs

### Phase 5: Fallback & Retry
- [ ] Implement fallback logic (open-sse/services/accountFallback.js)
- [ ] Implement error parsing (open-sse/utils/error.js)
- [ ] Implement retry logic with exponential backoff
- [ ] Test fallback chains

### Phase 6: Provider Executors
- [ ] Implement base executor (open-sse/executors/base.js)
- [ ] Implement 17 provider executors
- [ ] Implement OAuth token refresh
- [ ] Implement rate limit detection

### Phase 7: Quota Tracking
- [ ] Implement usage database (src/lib/usageDb.js)
- [ ] Implement token counting
- [ ] Implement cost calculation
- [ ] Implement quota reset tracking

### Phase 8: Dashboard & APIs
- [ ] Implement management APIs (/api/*)
- [ ] Implement settings APIs
- [ ] Implement usage analytics
- [ ] Embed Next.js frontend

---

## 🔗 Cross-References

### By Feature
- **RTK**: 9ROUTER_RTK_CAVEMAN.md + `_ref/9router/open-sse/rtk/`
- **Fallback**: 9ROUTER_REFERENCE.md + `_ref/9router/open-sse/services/accountFallback.js`
- **Translation**: 9ROUTER_REFERENCE.md + `_ref/9router/open-sse/translator/`
- **Executors**: 9ROUTER_CODEBASE_MAP.md + `_ref/9router/open-sse/executors/`
- **Quota**: 9ROUTER_REFERENCE.md + `_ref/9router/src/lib/usageDb.js`

### By Layer
- **API**: 9ROUTER_REFERENCE.md → API Endpoints + `_ref/9router/src/app/api/`
- **Core**: 9ROUTER_CODEBASE_MAP.md → Core Components + `_ref/9router/open-sse/`
- **Data**: 9ROUTER_REFERENCE.md → Data Model + `_ref/9router/src/lib/`
- **Frontend**: 9ROUTER_REFERENCE.md → Technology Stack + `_ref/9router/src/app/`

---

## 📞 Support & Community

- **GitHub Issues**: https://github.com/decolua/9router/issues
- **GitHub Discussions**: https://github.com/decolua/9router/discussions
- **Website**: https://9router.com
- **NPM**: https://www.npmjs.com/package/9router

---

## 📝 Document Index

| Document | Purpose | Audience | Length |
|----------|---------|----------|--------|
| 9ROUTER_REFERENCE.md | Comprehensive overview | Everyone | ~2000 lines |
| 9ROUTER_CODEBASE_MAP.md | Code structure & navigation | Developers | ~800 lines |
| 9ROUTER_RTK_CAVEMAN.md | Token compression deep dive | RTK implementers | ~600 lines |
| 9ROUTER_RESOURCES.md | This file - navigation guide | Everyone | ~400 lines |

---

## ✅ Verification Checklist

Before starting implementation:

- [ ] Read all 3 documentation files
- [ ] Explored `_ref/9router/` directory
- [ ] Understood RTK compression concept
- [ ] Understood 3-tier fallback routing
- [ ] Understood format translation
- [ ] Identified critical files to port
- [ ] Reviewed provider executors
- [ ] Checked AGENTS.md for current phase
- [ ] Confirmed Go module setup
- [ ] Confirmed chi router setup

---

**Last Updated**: May 2026  
**9Router Version**: v0.4.29  
**Status**: Complete Reference Package ✓
