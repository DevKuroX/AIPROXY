# AIPROXY Migration Parity Checklist

> **Purpose**: Behavior verification against 9router reference
> **Usage**: Verify after each phase
> **Reference**: 9router is the behavioral source of truth

---

# CRITICAL RULE

```
Behavior parity REQUIRED.
Architecture parity NOT required.
```

---

# Authentication Parity

## Login Flow

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Email/password form | ✓ | | `[ ]` |
| Invalid credential error inline | ✓ | | `[ ]` |
| Redirect after login | ✓ | | `[ ]` |
| Redirect target preserved (`?next=`) | ✓ | | `[ ]` |
| Token stored | ✓ | | `[ ]` |
| Session persists on reload | ✓ | | `[ ]` |

## OAuth Flow

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| OAuth start redirect | ✓ | | `[ ]` |
| Callback receives code | ✓ | | `[ ]` |
| Token stored after callback | ✓ | | `[ ]` |
| Redirect to dashboard | ✓ | | `[ ]` |

## Logout Flow

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Token cleared | ✓ | | `[ ]` |
| User state cleared | ✓ | | `[ ]` |
| Redirect to login | ✓ | | `[ ]` |
| No orphan state | ✓ | | `[ ]` |

---

# Streaming Parity

## Chat Stream

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Chunks render incrementally | ✓ | | `[ ]` |
| Chunk order preserved | ✓ | | `[ ]` |
| `done` event terminates | ✓ | | `[ ]` |
| Error event shows error | ✓ | | `[ ]` |
| Abort mid-stream works | ✓ | | `[ ]` |
| Network error handled | ✓ | | `[ ]` |
| Provider 5xx shows error | ✓ | | `[ ]` |
| Provider auth fail shows error | ✓ | | `[ ]` |

## Retry Behavior

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Retry indicator shows | ✓ | | `[ ]` |
| Retry handled by backend | ✓ | | `[ ]` |
| Frontend does NOT retry | ✓ | | `[ ]` |

---

# Provider Parity

## Provider List

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| List displays | ✓ | | `[ ]` |
| Sorting works | ✓ | | `[ ]` |
| Filtering works | ✓ | | `[ ]` |
| Badges display | ✓ | | `[ ]` |
| Loading state shows | ✓ | | `[ ]` |

## Provider CRUD

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Create provider | ✓ | | `[ ]` |
| Update provider | ✓ | | `[ ]` |
| Delete provider | ✓ | | `[ ]` |
| Reorder providers | ✓ | | `[ ]` |
| Validation errors inline | ✓ | | `[ ]` |

---

# Keys Parity

## API Keys

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| List displays | ✓ | | `[ ]` |
| Create key | ✓ | | `[ ]` |
| Update key | ✓ | | `[ ]` |
| Delete key | ✓ | | `[ ]` |
| Key validation | ✓ | | `[ ]` |

---

# Settings Parity

## Settings Save

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Settings form displays | ✓ | | `[ ]` |
| Save persists | ✓ | | `[ ]` |
| Validation errors inline | ✓ | | `[ ]` |
| Theme toggle works | ✓ | | `[ ]` |

---

# Usage Parity

## Usage Charts

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Charts render | ✓ | | `[ ]` |
| Data accurate | ✓ | | `[ ]` |
| Date range works | ✓ | | `[ ]` |
| Loading state shows | ✓ | | `[ ]` |

---

# Loading States Parity

| Component | 9router | AIPROXY | Status |
|-----------|---------|---------|--------|
| Dashboard loading | ✓ | | `[ ]` |
| Provider list loading | ✓ | | `[ ]` |
| Keys list loading | ✓ | | `[ ]` |
| Settings loading | ✓ | | `[ ]` |
| Usage charts loading | ✓ | | `[ ]` |
| Chat stream loading | ✓ | | `[ ]` |

---

# Error States Parity

| Error Type | 9router | AIPROXY | Status |
|------------|---------|---------|--------|
| Network error | ✓ | | `[ ]` |
| Auth error redirect | ✓ | | `[ ]` |
| Validation error inline | ✓ | | `[ ]` |
| Provider error shows | ✓ | | `[ ]` |
| Stream error shows | ✓ | | `[ ]` |

---

# Edge Cases Parity

## Stream Abort

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| User can abort stream | ✓ | | `[ ]` |
| Abort cleans up connection | ✓ | | `[ ]` |
| UI shows aborted state | ✓ | | `[ ]` |
| No hanging connections | ✓ | | `[ ]` |

## Long Response

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Long stream continues | ✓ | | `[ ]` |
| No timeout | ✓ | | `[ ]` |
| Chunks arrive in order | ✓ | | `[ ]` |

## Network Interruption

| Behavior | 9router | AIPROXY | Status |
|----------|---------|---------|--------|
| Error propagates | ✓ | | `[ ]` |
| User can retry | ✓ | | `[ ]` |
| No silent failure | ✓ | | `[ ]` |

---

# Verification Protocol

## After Each Phase

1. Run parity checklist for affected features
2. Compare behavior with 9router if unsure
3. Document any intentional behavior changes
4. Document any discovered behavior differences

## Phase-Specific Verification

| Phase | Verify These Features |
|-------|----------------------|
| P3 | Auth, Providers, Keys, Settings |
| P4 | v1/v1beta streaming |
| P5 | All auth flows |
| P6 | Usage charts, settings |
| P7 | All streaming |
| P8 | All features that used SQLite |
| P9 | Homepage, layout |
| P11 | Full parity sweep |

---

# Parity Decision Log

If behavior intentionally differs from 9router:

```markdown
## <Feature>: <Behavior>

**Decision Date**: YYYY-MM-DD
**Decision Maker**: [Name]

**9router Behavior**: [What 9router does]
**AIPROXY Behavior**: [What AIPROXY does]
**Reason**: [Why this difference is intentional]
```

---

# Sign-off

After Phase 11, this checklist must be fully verified:

```
AUTH PARITY: [ ] VERIFIED
STREAM PARITY: [ ] VERIFIED
PROVIDER PARITY: [ ] VERIFIED
KEYS PARITY: [ ] VERIFIED
SETTINGS PARITY: [ ] VERIFIED
USAGE PARITY: [ ] VERIFIED
LOADING PARITY: [ ] VERIFIED
ERROR PARITY: [ ] VERIFIED
EDGE CASE PARITY: [ ] VERIFIED

FINAL PARITY SIGN-OFF: ________________ Date: ________
```
