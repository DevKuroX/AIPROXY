# Phase 5 — Auth + Token Source Unification

**Objective**: Single token source of truth via `src/lib/api.ts`. Preserve 9router auth UX.

**Branch**: `phase/5-auth-unification`

**Dependencies**: Phase 1

**Exit Gate**: Manual auth regression suite passes; one token storage path.

---

## Tasks

| ID | Task | Target | Risk |
|----|------|--------|------|
| [T5.1](T5.1-token-paths-audit.md) | Token paths audit | `inventory/token-paths.md` | 🟢 |
| [T5.2](T5.2-token-decision-lock.md) | Token storage decision lock | doc only | 🟢 |
| [T5.3](T5.3-single-token-accessor.md) | Single token accessor | `src/lib/api.ts` | 🟠 |
| [T5.4](T5.4-userstore-refactor.md) | UserStore refactor | `src/store/userStore.js` | 🟠 |
| [T5.5](T5.5-auth-fetch-unification.md) | Auth fetch unification | `src/lib/*-api.ts` | 🟡 |
| [T5.6](T5.6-login-flow-audit.md) | Login flow audit | `src/app/login/*` | 🟠 |
| [T5.7](T5.7-oauth-callback-audit.md) | OAuth callback audit | `src/app/callback/*` | 🟠 |
| [T5.8](T5.8-logout-flow-audit.md) | Logout flow audit | logout buttons | 🟡 |
| [T5.9](T5.9-auth-regression-manual-test.md) | Auth regression manual test | regression doc | 🔴 |
| [T5.10](T5.10-direct-token-accessors.md) | Direct token accessors refactor | scattered | 🟠 |

---

## Auth Regression Scenarios

1. Login with valid credentials
2. Login with invalid credentials
3. Session persists on reload
4. Logout clears state
5. Expired token redirects
6. OAuth flow complete

## Forbidden

- Switching to cookies without PM approval
- Decoding JWT in frontend
- Storing tokens in Zustand userStore

## Exit Criteria

- [ ] Single source of truth: `src/lib/api.ts` token helpers
- [ ] `userStore` has no token state
- [ ] Manual auth regression suite passes
- [ ] Build green
