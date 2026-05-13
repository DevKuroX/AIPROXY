# AIPROXY API Contract

Backend base URL:

```env
http://localhost:1432
```

Frontend MUST follow these contracts exactly.

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

---

# Error Format

All backend errors follow:
```json
{
  "error": "message"
}
```

---

# Rules

- Backend is authoritative
- Frontend adapts to backend
- No frontend-only schemas
