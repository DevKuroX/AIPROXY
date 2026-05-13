# Development Guide

This guide covers setting up a development environment and contributing to AI Proxy.

## Prerequisites

- Go 1.25+
- Node.js 18+ (for frontend development)
- PostgreSQL 14+ (or use Docker)
- Make

## Quick Start

```bash
# Clone the repository
git clone https://github.com/DevKuroX/AIPROXY.git
cd AIPROXY

# Install Go dependencies
go mod download

# Set up environment
cp .env.example .env
# Edit .env with your local settings

# Create database
createdb ai_proxy

# Run migrations
make migrate

# Run the server
make run
```

## Project Structure

```
ai_proxy/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
│
├── internal/                     # Private application code
│   ├── api/                      # HTTP layer
│   │   ├── v1/                   # 0penAI-compatible endpoints
│   │   │   ├── chat.go           # POST /v1/chat/completions
│   │   │   ├── models.go         # GET /v1/models
│   │   │   ├── embeddings.go     # POST /v1/embeddings
│   │   │   └── routes.go         # Route registration
│   │   ├── v1beta/               # Gemini-compatible endpoints
│   │   └── admin/                # Dashboard/management API
│   │       ├── auth.go           # Login/logout
│   │       ├── providers.go      # Provider CRUD
│   │       ├── keys.go           # API key management
│   │       ├── aliases.go        # Model aliases
│   │       ├── nodes.go          # Custom provider nodes
│   │       ├── usage.go          # Usage statistics
│   │       └── cli.go            # CLI config helpers
│   │
│   ├── config/                   # Configuration loading
│   │   └── config.go
│   │
│   ├── executor/                 # Provider adapters
│   │   ├── base.go               # Executor interface
│   │   ├── default.go            # Standard HTTP executor
│   │   ├── registry.go           # Provider registry
│   │   ├── codex.go              # OpenAI Codex
│   │   ├── github.go             # GitHub Models
│   │   ├── opencode.go           # OpenCode
│   │   ├── gemini_cli.go         # Gemini CLI
│   │   ├── cursor.go             # Cursor
│   │   ├── kiro.go               # Kiro
│   │   └── ...                   # 30+ more providers
│   │
│   ├── router/                   # Core routing logic
│   │   ├── handler.go            # Main request handler
│   │   ├── resolver.go           # Model alias resolution
│   │   ├── fallback.go           # Fallback chain
│   │   ├── usage.go              # Usage tracking
│   │   └── usage_recorder.go     # Usage persistence
│   │
│   ├── rtk/                      # Token optimization
│   │   ├── caveman.go            # Caveman prompt injection
│   │   ├── autodetect.go         # Auto-detect RTK level
│   │   ├── apply.go              # Apply transformations
│   │   └── filters/              # Content filters
│   │       ├── dedup_log.go      # Deduplicate log lines
│   │       ├── gitdiff.go        # Git diff optimization
│   │       ├── grep.go           # Grep optimization
│   │       ├── tree.go           # Tree output optimization
│   │       └── ...
│   │
│   ├── storage/                  # Database layer
│   │   ├── db.go                 # Connection management
│   │   ├── migrations/           # SQL migrations
│   │   ├── providers.go          # Provider storage
│   │   ├── keys.go               # API key storage
│   │   ├── aliases.go            # Alias storage
│   │   ├── nodes.go              # Node storage
│   │   ├── usage.go              # Usage storage
│   │   └── pricing.go            # Pricing data
│   │
│   ├── translator/               # Format conversion
│   │   ├── types.go              # Shared types
│   │   ├── openai_to_claude.go   # 0penAI → CL4ude
│   │   ├── openai_to_gemini.go   # 0penAI → Gemini
│   │   ├── claude_to_openai.go   # CL4ude → 0penAI
│   │   └── gemini_to_openai.go   # Gemini → 0penAI
│   │
│   ├── models/                   # Data models
│   │   ├── provider.go
│   │   ├── key.go
│   │   ├── alias.go
│   │   ├── node.go
│   │   ├── usage.go
│   │   └── analytics.go
│   │
│   ├── pricing/                  # Cost calculation
│   │   ├── calculator.go         # Cost calculator
│   │   └── default_pricing.go    # Pricing data
│   │
│   ├── stream/                   # SSE streaming
│   │   └── writer.go
│   │
│   └── errs/                     # Error handling
│       └── errors.go
│
├── frontend/                     # Next.js dashboard
│   ├── src/
│   │   ├── app/
│   │   │   ├── login/
│   │   │   └── dashboard/
│   │   │       ├── providers/
│   │   │       ├── keys/
│   │   │       ├── aliases/
│   │   │       ├── nodes/
│   │   │       ├── combos/
│   │   │       ├── oauth/
│   │   │       └── analytics/
│   │   └── lib/
│   │       └── api.ts
│   ├── package.json
│   └── next.config.ts
│
├── docs/                         # Documentation
├── task/                         # Task specifications
├── result/                       # Phase results
└── Makefile
```

## Development Commands

```bash
# Run server
make run

# Build binary
make build

# Run tests
make test

# Run linter
make lint

# Run with live reload (requires air)
air

# Generate migrations
make migrate-create name=add_new_table

# Run migrations
make migrate

# Rollback last migration
make migrate-rollback
```

## Database Development

### Migrations

Migrations are SQL files in `internal/storage/migrations/`:

```
internal/storage/migrations/
├── 001_initial.sql
├── 002_api_keys.sql
├── 003_aliases.sql
├── 004_usage.sql
├── 005_pricing.sql
└── 006_nodes_aliases.sql
```

Create a new migration:

```bash
# Create file manually with sequential number
echo "-- Migration: add_new_table
CREATE TABLE new_table (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);
" > internal/storage/migrations/007_new_table.sql
```

### Database Queries

Use `pgx` for database access:

```go
// internal/storage/example.go
package storage

import "github.com/jackc/pgx/v5"

func GetExample(db *pgx.Conn, id int) (*Example, error) {
    const query = `SELECT id, name FROM examples WHERE id = $1`
    var e Example
    err := db.QueryRow(context.Background(), query, id).Scan(&e.ID, &e.Name)
    if err != nil {
        return nil, err
    }
    return &e, nil
}
```

## Adding a New Provider

### 1. Create Executor

```go
// internal/executor/new_provider.go
package executor

type NewProviderExecutor struct {
    *DefaultExecutor
    apiKey string
}

func NewNewProviderExecutor(apiKey string) *NewProviderExecutor {
    return &NewProviderExecutor{
        DefaultExecutor: NewDefaultExecutor("https://api.newprovider.com/v1", apiKey),
        apiKey: apiKey,
    }
}

// Override methods if needed
func (e *NewProviderExecutor) TransformRequest(req *ChatCompletionRequest) error {
    // Custom transformation logic
    return nil
}
```

### 2. Register in `registry.go`

```go
func init() {
    // ... existing registrations
    Register("new-provider", NewNewProviderExecutor(""))
}
```

### 3. Add OAuth Support (if needed)

```go
// internal/auth/oauth/new_provider.go
package oauth

func init() {
    Register("new-provider", &OAuthConfig{
        ClientID:     os.Getenv("NEW_PROVIDER_CLIENT_ID"),
        ClientSecret: os.Getenv("NEW_PROVIDER_CLIENT_SECRET"),
        AuthURL:      "https://api.newprovider.com/oauth/authorize",
        TokenURL:     "https://api.newprovider.com/oauth/token",
        DeviceURL:    "https://api.newprovider.com/oauth/device",
    })
}
```

### 4. Add Tests

```go
// internal/executor/new_provider_test.go
package executor

import "testing"

func TestNewProviderExecutor(t *testing.T) {
    exec := NewNewProviderExecutor("test-key")
    if exec == nil {
        t.Fatal("expected non-nil executor")
    }
}
```

## Adding a New RTK Filter

### 1. Create Filter

```go
// internal/rtk/filters/new_filter.go
package filters

func init() {
    Register("new-filter", NewFilter)
}

func NewFilter(input string) (string, error) {
    // Transformation logic
    return transformed, nil
}
```

### 2. Add Tests

```go
// internal/rtk/filters/new_filter_test.go
package filters

import "testing"

func TestNewFilter(t *testing.T) {
    input := "test input"
    output, err := NewFilter(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    // assertions
}
```

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/executor/...

# Run specific test
go test -run TestNewProviderExecutor ./internal/executor/
```

### Integration Tests

```bash
# Requires running PostgreSQL
DATABASE_URL="postgres://test:test@localhost:5432/ai_proxy_test?sslmode=disable" \
    go test -tags=integration ./...
```

### Frontend Tests

```bash
cd frontend

# Run tests
npm test

# Run e2e tests
npm run test:e2e
```

## Code Style

### Go

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable names
- Add comments for exported functions
- Handle errors explicitly (no ignored errors)

```go
// Good
func GetUser(id int) (*User, error) {
    user, err := db.GetUser(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return user, nil
}

// Bad
func GetUser(id int) *User {
    user, _ := db.GetUser(id)  // Ignoring error!
    return user
}
```

### TypeScript/React

- Use TypeScript strict mode
- Prefer functional components with hooks
- Use `async/await` over promise chains
- Keep components small and focused

```typescript
// Good
export async function fetchProviders(): Promise<Provider[]> {
  const response = await fetch('/api/providers');
  if (!response.ok) {
    throw new Error(`Failed to fetch: ${response.status}`);
  }
  return response.json();
}

// Bad
export function fetchProviders(): Promise<Provider[]> {
  return fetch('/api/providers').then(r => r.json());
}
```

## Debugging

### Enable Debug Logging

```bash
# Set log level via environment
LOG_LEVEL=debug make run
```

### Database Debugging

```bash
# Connect to database
psql $DATABASE_URL

# Check tables
\dt

# Check recent usage
SELECT * FROM usage ORDER BY created_at DESC LIMIT 10;
```

### Request Tracing

Add request ID to track requests:

```bash
curl -H "X-Request-ID: trace-123" http://localhost:20128/v1/chat/completions ...
```

## Making Changes

### Before Committing

1. Run tests: `make test`
2. Run linter: `make lint`
3. Format code: `gofmt -w .`
4. Update documentation if needed

### Commit Messages

Use conventional commits:

```
feat: add support for new provider
fix: correct token counting for streaming
docs: update deployment guide
refactor: simplify fallback chain logic
test: add tests for RTK filters
```

## Architecture Decisions

### Why Go?

- Single binary deployment
- Excellent HTTP server performance
- Strong typing and compile-time checks
- Great standard library
- Zero external runtime dependencies

### Why PostgreSQL?

- Robust and battle-tested
- Excellent JSON support for flexible schemas
- Great performance for read-heavy workloads
- Strong ecosystem (pgx driver)

### Why Separate Translators?

Each AI provider has slightly different request/response formats. Translators isolate this complexity:

```
Client → [0penAI format] → Router → Translator → [Provider format] → Provider
                                                              ↓
Client ← [0penAI format] ← Router ← Translator ← [Provider format] ←
```

## Getting Help

- Open an issue on GitHub
- Check existing documentation in `/docs`
- Review the reference implementation in `_ref/9router/`
