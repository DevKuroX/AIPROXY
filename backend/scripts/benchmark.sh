#!/bin/bash
set -e

BASE_URL="${BASE_URL:-http://localhost:20128}"
API_KEY="${API_KEY:-test-key}"
DURATION="${DURATION:-30s}"
CONCURRENT="${CONCURRENT:-50}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

check_dependencies() {
    if ! command -v hey &> /dev/null && ! command -v wrk &> /dev/null; then
        log_error "Neither 'hey' nor 'wrk' found. Install one:"
        echo "  hey: go install github.com/rakyll/hey@latest"
        echo "  wrk: https://github.com/wg/wrk"
        exit 1
    fi
}

wait_for_server() {
    log_info "Checking server availability at $BASE_URL..."
    for i in {1..30}; do
        if curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health" 2>/dev/null | grep -q "200\|404"; then
            log_info "Server is responding"
            return 0
        fi
        sleep 1
    done
    log_warn "Server health check timed out, proceeding anyway..."
}

run_hey_benchmark() {
    log_info "Running benchmark with hey ($DURATION, $CONCURRENT workers)..."
    
    local payload='{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Say hello"}],"max_tokens":10,"stream":false}'
    
    hey -n 1000 -c "$CONCURRENT" -d "$payload" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_KEY" \
        -H "Accept: application/json" \
        "$BASE_URL/v1/chat/completions" 2>&1 | tee /tmp/benchmark_hey.txt
    
    echo ""
    extract_hey_metrics
}

run_wrk_benchmark() {
    log_info "Running benchmark with wrk ($DURATION, $CONCURRENT connections)..."
    
    wrk -t4 -c"$CONCURRENT" -d"$DURATION" \
        -s "$(dirname "$0")/wrk_chat.lua" \
        "$BASE_URL" 2>&1 | tee /tmp/benchmark_wrk.txt
    
    echo ""
    extract_wrk_metrics
}

extract_hey_metrics() {
    log_info "Extracting key metrics..."
    
    local req_sec=$(grep "Requests/sec:" /tmp/benchmark_hey.txt | awk '{print $2}')
    local latency_avg=$(grep -A1 "Latency distribution" /tmp/benchmark_hey.txt | grep "Avg" | awk '{print $2}')
    local latency_p99=$(grep "99%" /tmp/benchmark_hey.txt | awk '{print $2}')
    
    echo "=========================================="
    echo "BENCHMARK SUMMARY (hey)"
    echo "=========================================="
    echo "  Requests/sec:    $req_sec"
    echo "  Avg Latency:     $latency_avg"
    echo "  P99 Latency:     $latency_p99"
    echo "=========================================="
}

extract_wrk_metrics() {
    log_info "Extracting key metrics..."
    
    local req_sec=$(grep "Requests/sec:" /tmp/benchmark_wrk.txt | awk '{print $2}')
    local latency_avg=$(grep "Latency" /tmp/benchmark_wrk.txt | head -1 | awk '{print $2}')
    local latency_p99=$(grep "99%" /tmp/benchmark_wrk.txt | awk '{print $2}')
    
    echo "=========================================="
    echo "BENCHMARK SUMMARY (wrk)"
    echo "=========================================="
    echo "  Requests/sec:    $req_sec"
    echo "  Avg Latency:     $latency_avg"
    echo "  P99 Latency:     $latency_p99"
    echo "=========================================="
}

run_quick_benchmark() {
    log_info "Running quick curl-based benchmark..."
    
    local start=$(date +%s.%N)
    local success=0
    local total=20
    
    for i in $(seq 1 $total); do
        local code=$(curl -s -o /dev/null -w "%{http_code}" \
            -X POST "$BASE_URL/v1/chat/completions" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $API_KEY" \
            -d '{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Hi"}],"max_tokens":5,"stream":false}' \
            --connect-timeout 5 --max-time 30)
        
        if [ "$code" = "200" ] || [ "$code" = "201" ]; then
            ((success++))
        fi
    done
    
    local end=$(date +%s.%N)
    local duration=$(echo "$end - $start" | bc)
    local req_sec=$(echo "scale=2; $total / $duration" | bc)
    
    echo "=========================================="
    echo "QUICK BENCHMARK SUMMARY"
    echo "=========================================="
    echo "  Total requests:  $total"
    echo "  Successful:      $success"
    echo "  Duration:        ${duration}s"
    echo "  Requests/sec:    $req_sec"
    echo "=========================================="
}

main() {
    log_info "AI Proxy Benchmark Suite"
    log_info "Target: $BASE_URL"
    log_info "Duration: $DURATION, Concurrent: $CONCURRENT"
    echo ""
    
    check_dependencies
    wait_for_server
    
    if command -v hey &> /dev/null; then
        run_hey_benchmark
    elif command -v wrk &> /dev/null; then
        run_wrk_benchmark
    else
        run_quick_benchmark
    fi
    
    rm -f /tmp/benchmark_hey.txt /tmp/benchmark_wrk.txt
    log_info "Benchmark complete"
}

main "$@"
