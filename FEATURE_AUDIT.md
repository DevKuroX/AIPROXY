# Feature Audit

Purpose:
Continuously compare:
- 9router behavior
- AIPROXY implementation

Goal:
Prevent feature regressions during migration.

---

# Authentication Audit

9router:
- password-only login
- cookie session
- redirect after login

AIPROXY:
- backend auth migrated
- frontend parity pending

Missing:
- login redirect parity
- session restore parity

---

# Streaming Audit

9router:
- chunk order stable
- stream abort supported

AIPROXY:
- backend stream implemented

Pending:
- frontend abort parity
- retry parity

---

# Providers Audit

9router:
- provider grouping
- health badges
- model alias support

AIPROXY:
- backend implemented

Pending:
- frontend badge parity
- filter parity

---

# Current Regression Risks

High risk:
- auth flow mismatch
- streaming UX mismatch
- hidden provider edge cases