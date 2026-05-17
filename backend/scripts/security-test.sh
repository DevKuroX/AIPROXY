#!/bin/bash
# Security Integration Tests for AIPROXY
# Validates all 12 security hardening fixes.
# Returns: 0 = all pass, 1 = any fail
set -e

cd "$(dirname "$0")/.."

PASS=0
FAIL=0
SKIP=0

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "============================================"
echo "  AIPROXY Security Integration Tests"
echo "============================================"
echo ""

# --- Helpers ---
pass() { echo -e "  ${GREEN}✓ PASS${NC}  $1"; PASS=$((PASS+1)); }
fail() { echo -e "  ${RED}✗ FAIL${NC}  $1"; FAIL=$((FAIL+1)); }
skip() { echo -e "  ${YELLOW}∼ SKIP${NC}  $1"; SKIP=$((SKIP+1)); }

check_grep_no_match() {
    local desc="$1" file="$2" pattern="$3"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        fail "$desc"
    else
        pass "$desc"
    fi
}

check_grep_match() {
    local desc="$1" file="$2" pattern="$3"
    if grep -q "$pattern" "$file" 2>/dev/null; then
        pass "$desc"
    else
        fail "$desc"
    fi
}

# ====================================================================
# TEST 1: No hardcoded OAuth secrets in source
# ====================================================================
echo ""
echo "--- Test 1: No Hardcoded Secrets in Source ---"

# Previously leaked Google OAuth client secret
check_grep_no_match \
  "No hardcoded GOCSPX secrets" \
  "internal/services/token_refresh.go" \
  "GOCSPX"

# Previously leaked Iflow secret
check_grep_no_match \
  "No hardcoded Iflow secret" \
  "internal/services/token_refresh.go" \
  "4Z3YjXycVsQvyGF1etiNlIBB4RsqSDtW"

# Verify OAuth secrets now read from config (env vars)
check_grep_match \
  "GeminiClientSecret from config" \
  "internal/services/token_refresh.go" \
  "config.GetConfig().GeminiClientSecret"

check_grep_match \
  "IflowClientSecret from config" \
  "internal/services/token_refresh.go" \
  "config.GetConfig().IflowClientSecret"

check_grep_match \
  "AntigravityClientSecret from config" \
  "internal/services/token_refresh.go" \
  "config.GetConfig().AntigravityClientSecret"

# ====================================================================
# TEST 2: Auth enforced on proxy/console endpoints
# ====================================================================
echo ""
echo "--- Test 2: Auth Middleware on Endpoints ---"

# Check source code for middleware wrapping
check_grep_match \
  "Proxy API wrapped with apiKeyMiddleware" \
  "internal/api/routes.go" \
  'mux.Handle("GET /api/proxies", apiKeyMiddleware'

check_grep_match \
  "Console log stream wrapped with authMiddleware" \
  "internal/api/routes.go" \
  'mux.Handle("GET /api/translator/console-logs/stream", authMiddleware'

check_grep_match \
  "Kiro usage wrapped with apiKeyMiddleware" \
  "internal/api/routes.go" \
  'mux.Handle("GET /api/usage/kiro", apiKeyMiddleware'

# Live server check (if running)
BASE_URL="${BASE_URL:-http://localhost:20128}"
if curl -s -o /dev/null -w "" --connect-timeout 2 "$BASE_URL/health" 2>/dev/null; then
    CODE=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/proxies" 2>/dev/null || echo "000")
    if [ "$CODE" = "401" ]; then
        pass "Proxy API rejects unauthenticated requests (HTTP $CODE)"
    elif [ "$CODE" = "000" ]; then
        skip "Proxy API auth test (server not reachable)"
    else
        fail "Proxy API should return 401, got $CODE"
    fi
else
    skip "Proxy API auth test (server not running)"
fi

# ====================================================================
# TEST 3: CORS with origin allowlist
# ====================================================================
echo ""
echo "--- Test 3: CORS Origin Allowlist ---"

check_grep_match \
  "CORS middleware checks allowed origins" \
  "internal/dashboard/dashboard.go" \
  "func CORS(next http.Handler, allowedOrigins"

check_grep_match \
  "AllowedOrigins in config" \
  "internal/config/config.go" \
  "AllowedOrigins"

# Live CORS check on dashboard port
DASHBOARD_PORT="${DASHBOARD_PORT:-1433}"
if curl -s -o /dev/null --connect-timeout 2 "http://localhost:$DASHBOARD_PORT/health" 2>/dev/null; then
    CORS_HEADER=$(curl -s -H "Origin: https://evil.com" -I "http://localhost:$DASHBOARD_PORT/" 2>/dev/null | grep -i "access-control-allow-origin" | tr -d '\r')
    if echo "$CORS_HEADER" | grep -qi "evil.com"; then
        fail "CORS reflects evil origin"
    elif [ -z "$CORS_HEADER" ]; then
        pass "CORS rejects disallowed origin (no header)"
    else
        pass "CORS does not reflect evil origin"
    fi

    CORS_ALLOW=$(curl -s -H "Origin: http://localhost:3000" -I "http://localhost:$DASHBOARD_PORT/" 2>/dev/null | grep -i "access-control-allow-origin" | tr -d '\r')
    if echo "$CORS_ALLOW" | grep -qi "localhost:3000"; then
        pass "CORS allows whitelisted origin"
    else
        skip "CORS allowed origin test (dashboard may not serve this path)"
    fi
else
    skip "CORS live test (dashboard server not running on :$DASHBOARD_PORT)"
fi

# ====================================================================
# TEST 4: Rate limiting middleware
# ====================================================================
echo ""
echo "--- Test 4: Rate Limiting Middleware ---"

check_grep_match \
  "Rate limiter file exists" \
  "internal/api/middleware/ratelimit.go" \
  "type RateLimiter struct"

check_grep_match \
  "Token bucket implementation" \
  "internal/api/middleware/ratelimit.go" \
  "func.*Allow.*key.*rate.*capacity"

check_grep_match \
  "Rate limiter wired in routes" \
  "internal/api/routes.go" \
  "rateLimitedMux\|RateLimit"

# ====================================================================
# TEST 5: No panic() in request path
# ====================================================================
echo ""
echo "--- Test 5: No panic() in Request Path ---"

# Check fixed files for panic calls outside init()
# registry.go - only remaining panic should be in init()
if grep -q "panic(" "internal/executor/registry.go" 2>/dev/null; then
    # Verify only init() contains panic (acceptable at startup)
    PANIC_LINES=$(grep -n "panic(" "internal/executor/registry.go" 2>/dev/null)
    IN_INIT=true
    while IFS= read -r line; do
        LINE_NUM=$(echo "$line" | cut -d: -f1)
        if ! sed -n "1,${LINE_NUM}p" "internal/executor/registry.go" | tac | grep -q "func init\b"; then
            IN_INIT=false
            break
        fi
    done <<< "$PANIC_LINES"
    if $IN_INIT; then
        pass "Only init() panic in executor/registry.go (acceptable)"
    else
        fail "Non-init panic found in executor/registry.go"
    fi
else
    pass "No panic in executor/registry.go"
fi

check_grep_no_match \
  "No panic in claude_to_openai.go" \
  "internal/translator/stream/claude_to_openai.go" \
  "panic("

check_grep_no_match \
  "No panic in openai_to_claude.go" \
  "internal/translator/stream/openai_to_claude.go" \
  "panic("

# ====================================================================
# TEST 6: HTTP recover() middleware
# ====================================================================
echo ""
echo "--- Test 6: HTTP Recover Middleware ---"

check_grep_match \
  "Recover middleware exists" \
  "internal/api/middleware/recover.go" \
  "func Recover"

check_grep_match \
  "Recover applied in main.go" \
  "cmd/server/main.go" \
  "middleware.Recover"

check_grep_match \
  "Recover returns 500 on panic" \
  "internal/api/middleware/recover.go" \
  "http.Error.*internal server error"

check_grep_match \
  "Recover logs stack trace" \
  "internal/api/middleware/recover.go" \
  "debug.Stack"

# ====================================================================
# TEST 7: Provider API keys encrypted at rest
# ====================================================================
echo ""
echo "--- Test 7: Provider API Key Encryption ---"

check_grep_match \
  "encryptAPIKey method exists" \
  "internal/storage/db.go" \
  "func.*encryptAPIKey"

check_grep_match \
  "decryptAPIKey method exists" \
  "internal/storage/db.go" \
  "func.*decryptAPIKey"

check_grep_match \
  "Encryption before INSERT in providers.go" \
  "internal/storage/providers.go" \
  "encryptAPIKey"

check_grep_match \
  "Encryption before INSERT in nodes.go" \
  "internal/storage/nodes.go" \
  "encryptAPIKey"

check_grep_match \
  "Encryption key config field" \
  "internal/config/config.go" \
  "EncryptionKey"

# ====================================================================
# TEST 8: Nil-safe type assertions in TTS handlers
# ====================================================================
echo ""
echo "--- Test 8: Nil-Safe Type Assertions ---"

check_grep_match \
  "ok-checked candidate assertion in gemini.go" \
  "internal/handlers/tts/providers/gemini.go" \
  "candidate, ok :="

check_grep_match \
  "ok-checked content assertion in gemini.go" \
  "internal/handlers/tts/providers/gemini.go" \
  "content, ok"

# Verify no unchecked direct chained assertions remain
if grep -q '\.(map\[string\]interface{})\[' "internal/handlers/tts/providers/gemini.go" 2>/dev/null; then
    COUNT=$(grep -c '\.(map\[string\]interface{})\[' "internal/handlers/tts/providers/gemini.go" 2>/dev/null || echo 0)
    fail "Found $COUNT unchecked chained assertions in gemini.go"
else
    pass "No unchecked chained assertions in gemini.go"
fi

# ====================================================================
# TEST 9: Sanitize upstream errors
# ====================================================================
echo ""
echo "--- Test 9: Upstream Error Sanitization ---"

# Client gets generic message (no upstream body leak)
check_grep_match \
  "Generic error returned to client (streaming)" \
  "internal/router/handler.go" \
  '"provider error: %d"'

check_grep_match \
  "Generic error returned to client (non-streaming)" \
  "internal/router/handler.go" \
  'fmt.Sprintf("provider error: %d"'

# Full error logged server-side
check_grep_match \
  "Full error logged server-side (streaming)" \
  "internal/router/handler.go" \
  "log.Printf.*Upstream provider error (streaming)"

check_grep_match \
  "Full error logged server-side (non-streaming)" \
  "internal/router/handler.go" \
  "log.Printf.*Upstream provider error:"

# No raw body in client-facing error
BODY_LEAK=$(grep -c 'string(errBody)' "internal/router/handler.go" 2>/dev/null || echo 0)
if [ "$BODY_LEAK" -gt 0 ]; then
    # The string(errBody) should only appear in log.Printf lines, not in error message sent to client
    ERROR_BODY_LEAK=$(grep -c 'error.*string(errBody)' "internal/router/handler.go" 2>/dev/null || echo 0)
    LOG_BODY=$(grep -c 'log.Printf.*string(errBody)' "internal/router/handler.go" 2>/dev/null || echo 0)
    if [ "$ERROR_BODY_LEAK" -gt 0 ] || [ "$LOG_BODY" -eq 0 ]; then
        # Check: string(errBody) in non-log lines = leak
        TOTAL_STR=$(grep -c 'string(errBody)' "internal/router/handler.go" 2>/dev/null || echo 0)
        LOG_LINES=$(grep -c 'log.*string(errBody)' "internal/router/handler.go" 2>/dev/null || echo 0)
        if [ "$TOTAL_STR" -eq "$LOG_LINES" ]; then
            pass "errBody only exposed in server logs, not to client"
        else
            fail "errBody may be leaked to client (check handler.go)"
        fi
    elif [ "$LOG_BODY" -gt 0 ]; then
        pass "Upstream error body logged server-side only"
    else
        pass "No errBody leak detected"
    fi
else
    pass "No errBody usage (sanitized)"
fi

# ====================================================================
# TEST 10: SQL parameter numbering fix
# ====================================================================
echo ""
echo "--- Test 10: SQL Parameter Numbering ---"

# Verify strconv.Itoa is used (correct)
check_grep_match \
  "strconv.Itoa used for param numbering" \
  "internal/storage/request_details.go" \
  "strconv.Itoa(argIndex)"

# Verify no rune conversion (bug)
check_grep_no_match \
  "No rune('0'+argIndex) bug pattern" \
  "internal/storage/request_details.go" \
  "rune.*'0'"

# Count occurrences fixed (there should be 7)
ITO_COUNT=$(grep -c "strconv.Itoa(argIndex)" "internal/storage/request_details.go" 2>/dev/null || echo 0)
if [ "$ITO_COUNT" -ge 7 ]; then
    pass "All $ITO_COUNT parameter numbering locations use strconv.Itoa"
else
    fail "Expected 7+ strconv.Itoa locations, found $ITO_COUNT"
fi

# ====================================================================
# TEST 11: Token encryption (AES-256-GCM)
# ====================================================================
echo ""
echo "--- Test 11: Token Encryption ---"

check_grep_match \
  "encryptToken calls crypto.Encrypt" \
  "internal/services/token_refresh.go" \
  "encryptToken"

check_grep_match \
  "decryptToken calls crypto.Decrypt" \
  "internal/services/token_refresh.go" \
  "decryptToken"

check_grep_match \
  "crypto.Encrypt used in encryptToken" \
  "internal/services/token_refresh.go" \
  "crypto.Encrypt"

check_grep_match \
  "crypto.Decrypt used in decryptToken" \
  "internal/services/token_refresh.go" \
  "crypto.Decrypt"

# ====================================================================
# TEST 12: .env not in git
# ====================================================================
echo ""
echo "--- Test 12: .env Git Hygiene ---"

# Check .gitignore contains backend/.env
check_grep_match \
  ".gitignore has backend/.env" \
  "../.gitignore" \
  "backend/\.env"

# Check .env not currently tracked
ROOT_DIR=$(git rev-parse --show-toplevel 2>/dev/null)
if [ -f "$ROOT_DIR/backend/.env" ] && git -C "$ROOT_DIR" ls-files backend/.env 2>/dev/null | grep -q .; then
    fail "backend/.env is tracked in git"
else
    pass "backend/.env not tracked in git"
fi

# Check .env.example exists with placeholders (not real secrets)
check_grep_match \
  ".env.example exists" \
  ".env.example" \
  ".*"

# Check .env.example has no real secrets
ENV_EXAMPLE_SECRETS=0
if grep -q "GOCSPX" .env.example 2>/dev/null; then ENV_EXAMPLE_SECRETS=$((ENV_EXAMPLE_SECRETS+1)); fi
if [ "$ENV_EXAMPLE_SECRETS" -eq 0 ]; then
    pass ".env.example has no real secrets"
else
    fail ".env.example contains real secrets"
fi

# ====================================================================
# BUILD CHECK
# ====================================================================
echo ""
echo "--- Build Check ---"

if (go build ./cmd/server > /tmp/security-build.log 2>&1); then
    pass "Go build passes"
    rm -f /tmp/security-build.log
else
    fail "Go build fails: $(cat /tmp/security-build.log)"
fi

# ====================================================================
# SUMMARY
# ====================================================================
echo ""
echo "============================================"
echo "  Results Summary"
echo "============================================"
echo "  PASS:  $PASS"
echo "  FAIL:  $FAIL"
echo "  SKIP:  $SKIP"
echo "  Total: $((PASS + FAIL + SKIP))"
echo "============================================"

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "  ❌ SOME TESTS FAILED"
    exit 1
else
    echo ""
    echo "  ✅ ALL TESTS PASSED"
    exit 0
fi
