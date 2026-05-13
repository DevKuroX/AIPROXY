# Contributing to AI Proxy

Thank you for your interest in contributing to AI Proxy! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)
- [Documentation](#documentation)

## Code of Conduct

Be respectful and inclusive. We welcome contributions from everyone. Harassment, discrimination, and offensive behavior are not tolerated.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/AIPROXY.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Submit a pull request

## Development Setup

See [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) for detailed setup instructions.

Quick setup:

```bash
# Install dependencies
go mod download

# Set up environment
cp .env.example .env

# Create database
createdb ai_proxy

# Run migrations
make migrate

# Start development server
make run
```

## How to Contribute

### Reporting Bugs

1. Check existing issues to avoid duplicates
2. Use the bug report template
3. Include:
   - Steps to reproduce
   - Expected behavior
   - Actual behavior
   - Environment details (OS, Go version, etc.)

### Suggesting Features

1. Check existing issues and discussions
2. Use the feature request template
3. Describe the use case and expected behavior

### Contributing Code

**Good first issues**: Look for issues labeled `good first issue` or `help wanted`.

Areas that often need contributions:
- New provider executors
- RTK filter improvements
- Documentation improvements
- Test coverage

## Pull Request Process

### Before Submitting

1. **Run tests**: `make test`
2. **Run linter**: `make lint`
3. **Format code**: `gofmt -w .`
4. **Update docs**: If your change affects user-facing behavior

### PR Guidelines

1. **One feature per PR**: Keep changes focused
2. **Small PRs**: Easier to review, faster to merge
3. **Descriptive title**: Summarize the change
4. **Fill out the template**: Link issues, describe changes

### PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Tests pass locally
- [ ] New tests added for new functionality

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
```

### Review Process

1. Automated checks must pass (tests, lint)
2. At least one maintainer approval required
3. Address review feedback promptly
4. Squash commits before merge (maintainer will handle)

## Code Style

### Go

- Use `gofmt` for formatting
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use meaningful variable names
- Add comments for exported functions

```go
// Good
// GetUser retrieves a user by ID from the database.
// Returns ErrNotFound if the user does not exist.
func GetUser(id int) (*User, error) {
    // ...
}

// Bad
func getUser(i int) *User {  // unexported, no comment, ignores error
    // ...
}
```

### Error Handling

Always handle errors explicitly:

```go
// Good
user, err := db.GetUser(id)
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// Bad
user, _ := db.GetUser(id)  // Ignored error
```

### Imports

Group imports:

```go
import (
    // Standard library
    "context"
    "fmt"

    // External packages
    "github.com/go-chi/chi/v5"

    // Internal packages
    "github.com/DevKuroX/AIPROXY/internal/models"
)
```

### TypeScript/React

- Use TypeScript strict mode
- Prefer functional components with hooks
- Use `async/await` over promise chains

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestName ./path/to/package
```

### Writing Tests

```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:  "basic case",
            input: "hello",
            want:  "HELLO",
        },
        {
            name:    "error case",
            input:   "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("NewFeature() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("NewFeature() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Documentation

### When to Update

Update documentation when:
- Adding new features
- Changing API behavior
- Fixing documentation bugs
- Improving clarity

### Documentation Files

| File | Purpose |
|------|---------|
| `README.md` | Project overview, quickstart |
| `docs/DEPLOYMENT.md` | Production deployment |
| `docs/DEVELOPMENT.md` | Development guide |
| `docs/API.md` | API reference |
| `docs/ARCHITECTURE.md` | System architecture |

### Style

- Use clear, concise language
- Include code examples
- Keep examples up-to-date with code

## Questions?

- Open a GitHub issue for questions
- Check existing documentation first
- Be specific about what you're trying to do

## Recognition

Contributors are recognized in:
- GitHub contributors list
- Release notes for significant contributions

Thank you for contributing!
