# AIPROXY API Contract

Backend base URL:

```env
http://localhost:1432
```

Frontend MUST follow these contracts exactly.

Do not invent:
- endpoints
- payload shapes
- response formats

Backend schemas are authoritative.

---

# Authentication

## GET /api/auth/me

Response:

```json
{
  "id": "string",
  "email": "string",
  "name": "string"
}
```

---

## POST /api/auth/login

Request:

```json
{
  "email": "string",
  "password": "string"
}
```

Response:

```json
{
  "token": "string",
  "user": {
    "id": "string",
    "email": "string"
  }
}
```

---

# Providers

## GET /api/providers

Response:

```json
{
  "providers": []
}
```

---

## POST /api/providers

Request:

```json
{
  "name": "string",
  "type": "string"
}
```

---

# Settings

## GET /api/settings

Response:

```json
{
  "theme": "dark"
}
```

---

## POST /api/settings

Request:

```json
{
  "theme": "dark"
}
```

---

# Usage

## GET /api/usage

Response:

```json
{
  "total_requests": 0,
  "total_tokens": 0
}
```

---

# Aliases

## GET /api/aliases

Response:

```json
{
  "aliases": []
}
```

---

# Combos

## GET /api/combos

Response:

```json
{
  "combos": []
}
```

---

# OAuth

## GET /api/oauth/providers

Response:

```json
{
  "providers": []
}
```

---

# Streaming

## POST /api/chat/stream

Request:

```json
{
  "messages": [],
  "model": "string",
  "stream": true
}
```

Response:
- SSE stream
- normalized backend chunks

Frontend MUST NOT parse provider-native stream formats.

---

# Error Format

All backend errors must follow:

```json
{
  "error": "message"
}
```

---

# API Rules

- Backend is authoritative
- Frontend adapts to backend
- No temporary response formats
- No frontend-only schemas
- No compatibility wrappers unless approved

---

# Frontend API Layer

Frontend communication must go through:

```txt
src/lib/api/*
```

Do not:
- duplicate fetch wrappers
- fetch directly inside components
- create alternate API layers

---

# Route Strategy

Current strategy:
- keep Next.js API routes temporarily
- convert them into thin proxies

Final target:
- no business logic in routes
- no persistence in routes