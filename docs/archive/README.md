# AI Proxy

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg)]()
[![Status](https://img.shields.io/badge/status-production-green.svg)]()

A high-performance AI gateway that provides a unified OpenAI-compatible API for 40+ AI providers with intelligent fallback, token optimization, and zero external runtime dependencies.

## Features

- **OpenAI-Compatible API** - Drop-in replacement for OpenAI's `/v1/chat/completions` endpoint
- **Multi-Provider Support** - 40+ providers including Claude, Gemini, GPT-4, Codex, GitHub Models, Grok, and more
- **Intelligent Fallback** - Automatic failover between providers (subscription → pay-per-use → free tiers)
- **Token Optimization (RTK)** - 20-40% token savings with Request Token Kompressor using Caveman prompts
- **OAuth Integration** - Device flow authentication for Claude, Gemini, Codex, GitHub, Kiro, Cursor, and more
- **Model Aliases** - Create custom model names that resolve to specific providers
- **Usage Analytics** - Track costs, tokens, and requests per API key, model, or time period
- **Single Binary Deployment** - Go backend with embedded Next.js dashboard, no Node.js runtime required

## Quickstart

### Prerequisites

- Ubuntu 20.04+ (or any Linux with Go 1.25+)
- PostgreSQL 14+ (or use SQLite for development)
- Go 1.25+

### Installation

```bash
# Clone the repository
git clone https://github.com/DevKuroX/AIPROXY.git
cd AIPROXY

# Install Go dependencies
go mod download

# Set up environment
cp .env.example .env
# Edit .env with your settings

# Create PostgreSQL database
sudo -u postgres createdb ai_proxy
# Or update DATABASE_URL in .env for your setup

# Run database migrations
make migrate

# Build and run
make build
./bin/ai_proxy
```

The server will start on port 20128 (configurable via `PORT` env var).

### First Request

```bash
# Health check
curl http://localhost:20128/health

# Create an API key (requires admin auth)
curl -X POST http://localhost:20128/api/keys \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<session-cookie>" \
  -d '{"name": "my-key"}'

# Chat completion
curl http://localhost:20128/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### Dashboard

Access the admin dashboard at `http://localhost:20128/dashboard`. Default credentials:
- Username: `admin`
- Password: `admin` (set via `ADMIN_PASSWORD` env var)

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/ai_proxy?sslmode=disable` | PostgreSQL connection string |
| `JWT_SECRET` | `change-me-in-production` | Secret for JWT token signing |
| `ADMIN_PASSWORD` | `admin` | Password for admin dashboard login |
| `PORT` | `20128` | Server port |

### Provider Configuration

Configure providers through the dashboard or API:

```bash
# Add a provider via API
curl -X POST http://localhost:20128/api/providers \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<session-cookie>" \
  -d '{
    "name": "openai-main",
    "type": "openai",
    "api_key": "sk-...",
    "base_url": "https://api.openai.com/v1",
    "models": ["gpt-4o", "gpt-4o-mini", "o1", "o3-mini"]
  }'
```

### Model Aliases

Create aliases for easier model selection:

```bash
curl -X POST http://localhost:20128/api/models/alias \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<session-cookie>" \
  -d '{
    "alias": "fast",
    "model": "gpt-4o-mini"
  }'
```

Now you can use `fast` as the model name in requests.

## API Documentation

### OpenAI-Compatible Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1/chat/completions` | Chat completions (streaming supported) |
| `POST` | `/v1/embeddings` | Generate embeddings |
| `GET` | `/v1/models` | List available models |

### Admin Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/auth/login` | Admin login |
| `GET` | `/api/providers` | List providers |
| `POST` | `/api/providers` | Create provider |
| `GET` | `/api/keys` | List API keys |
| `POST` | `/api/keys` | Create API key |
| `GET` | `/api/usage/summary` | Usage statistics |
| `GET` | `/api/oauth/:provider/start` | Start OAuth flow |

### Gemini-Compatible Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/v1beta/models/:model:generateContent` | Generate content |
| `POST` | `/v1beta/models/:model:streamGenerateContent` | Stream content |

For full API documentation, see [docs/API.md](docs/API.md).

## Architecture

```
ai_proxy/
├── cmd/server/           # Application entry point
├── internal/
│   ├── api/              # HTTP handlers
│   │   ├── v1/           # OpenAI-compatible API
│   │   ├── v1beta/       # Gemini-compatible API
│   │   └── admin/        # Dashboard API
│   ├── config/           # Configuration loading
│   ├── executor/         # Provider adapters (40+)
│   ├── router/           # Request routing, fallback, usage
│   ├── rtk/              # Token optimization
│   ├── storage/          # Database layer
│   └── translator/       # Format conversion
├── frontend/             # Next.js dashboard
└── docs/                 # Documentation
```

## Development

```bash
# Run in development mode
make run

# Run tests
make test

# Run linter
make lint

# Build for production
make build
```

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed development guide.

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for:
- Docker deployment
- PostgreSQL setup
- Reverse proxy configuration (nginx/caddy)
- Production hardening

## Documentation

- [API Reference](docs/API.md) - Endpoint documentation
- [Architecture](docs/ARCHITECTURE.md) - System design
- [Deployment Guide](docs/DEPLOYMENT.md) - Production setup
- [Development Guide](docs/DEVELOPMENT.md) - Contributing
- [RTK Specification](docs/RTK_SPEC.md) - Token optimization details
- [Executor Reference](docs/EXECUTORS.md) - Provider adapters

## License

MIT License - see [LICENSE](LICENSE) file.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.
