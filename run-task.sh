#!/bin/bash

# AIPROXY Migration Task Runner
# Usage: ./run-task.sh <task-id>
# 
# This script generates a prompt for AI agents to execute migration tasks.
# Each task is bounded, atomic, and follows strict execution rules.

set -e

TASK_ID="$1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

if [ -z "$TASK_ID" ]; then
    echo -e "${BLUE}AIPROXY Migration Task Runner${NC}"
    echo ""
    echo "Usage: ./run-task.sh <task-id>"
    echo ""
    echo "Examples:"
    echo "  ./run-task.sh T0.1    # SQLite import inventory"
    echo "  ./run-task.sh T2.5    # Create usage API client"
    echo "  ./run-task.sh T4.9    # v1 proxy streaming"
    echo "  ./run-task.sh T8.3    # Delete SQLite db directory"
    echo ""
    echo -e "${YELLOW}Available tasks by phase:${NC}"
    echo ""
    
    for phase_dir in tasks/phase-*/; do
        phase_name=$(basename "$phase_dir")
        phase_num=$(echo "$phase_name" | sed 's/phase-//')
        
        # Get phase title from README
        if [ -f "${phase_dir}README.md" ]; then
            phase_title=$(head -1 "${phase_dir}README.md" | sed 's/^# //')
        else
            phase_title="$phase_name"
        fi
        
        echo -e "${GREEN}Phase $phase_num${NC} - $phase_title"
        
        # List tasks in this phase
        find "$phase_dir" -name "T*.md" -type f | grep -v README | sort | while read task_file; do
            task_id=$(basename "$task_file" | sed 's/\.md$//')
            # Get task title from file
            task_title=$(grep "^# " "$task_file" | head -1 | sed 's/^# //')
            echo "  $task_id"
        done
        echo ""
    done
    
    echo -e "${BLUE}Total: $(find tasks -name "T*.md" -type f | grep -v README | wc -l) tasks${NC}"
    exit 1
fi

# Validate task ID format
if ! [[ "$TASK_ID" =~ ^T[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}ERROR: Invalid task ID format. Expected: T<phase>.<seq> (e.g., T2.5)${NC}"
    exit 1
fi

# Find task file
TASK_FILE=$(find tasks -name "${TASK_ID}-*.md" -type f | grep -v README | head -1)

if [ -z "$TASK_FILE" ]; then
    echo -e "${RED}ERROR: Task ${TASK_ID} not found${NC}"
    echo ""
    echo "Available tasks:"
    find tasks -name "T*.md" -type f | grep -v README | sed 's/tasks\//  /' | sed 's/-.*\.md//' | sort
    exit 1
fi

# Extract task title
TASK_TITLE=$(grep "^# " "$TASK_FILE" | head -1 | sed 's/^# //')

# Get phase number
PHASE_NUM=$(echo "$TASK_ID" | sed 's/T\([0-9]*\)\..*/\1/')
PHASE_DIR="tasks/phase-$(printf '%02d' $PHASE_NUM)"

# Check prerequisites
check_prerequisites() {
    local task_file="$1"
    local prereq_section=$(grep -A 20 "^## Prerequisites" "$task_file" 2>/dev/null | grep -E "^\-|T[0-9]" | head -10)
    
    if [ -n "$prereq_section" ]; then
        echo -e "${YELLOW}Prerequisites:${NC}"
        echo "$prereq_section" | while read line; do
            echo "  $line"
        done
        echo ""
    fi
}

# Check if task is already done in TASK_STATUS.md
check_task_status() {
    local task_id="$1"
    if [ -f "TASK_STATUS.md" ]; then
        status=$(grep "| $task_id " TASK_STATUS.md | grep -o '\`.\`' | head -1)
        case "$status" in
            '\`[x]\`') echo -e "${GREEN}Status: DONE${NC}" ;;
            '\`[~]\`') echo -e "${YELLOW}Status: IN PROGRESS${NC}" ;;
            '\`[!]\`') echo -e "${RED}Status: BLOCKED${NC}" ;;
            '\`[-]\`') echo -e "${BLUE}Status: SKIPPED${NC}" ;;
            *)        echo -e "Status: PENDING" ;;
        esac
    fi
}

echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}TASK: ${TASK_ID} - ${TASK_TITLE}${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""
echo -e "File: ${TASK_FILE}"
echo ""

# Show task status
check_task_status "$TASK_ID"
echo ""

# Show prerequisites
check_prerequisites "$TASK_FILE"

echo -e "${BLUE}----------------------------------------${NC}"
echo -e "${BLUE}AI EXECUTION PROMPT:${NC}"
echo -e "${BLUE}----------------------------------------${NC}"
echo ""

# Generate prompt
cat << 'PROMPT_HEADER'
Read and follow strictly:

- EXECUTION_LOOP.md
- BUILD_RULES.md
- AGENT_RULES.md
- PROJECT_CONSTRAINTS.md
- TASK_STATUS.md

Then execute ONLY this task:

PROMPT_HEADER

echo "${TASK_FILE}"

cat << 'PROMPT_FOOTER'

Rules:
- Do NOT continue to another task
- Do NOT modify unrelated files
- Do NOT redesign architecture
- Do NOT fix unrelated build errors
- Read FIRST root build error only
- Preserve 9router behavior parity
- Stop immediately after task completion
- Maximum delegation depth: 1
- Follow BUILD_RULES.md for any build failures

Required output:
1. Files modified
2. Changes made
3. Verification results (tsc, build, tests)
4. TASK_STATUS.md update
5. Any blockers

After completion:
- Update TASK_STATUS.md: mark task as [x]
- Commit with message: "<type>: <description> (${TASK_ID})"

STOP after task execution.
PROMPT_FOOTER

echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${YELLOW}Tip: Pipe this output to your AI assistant${NC}"
echo -e "${YELLOW}     ./run-task.sh T2.5 | pbcopy${NC}"
echo -e "${YELLOW}     ./run-task.sh T2.5 | xclip -selection clipboard${NC}"
echo -e "${BLUE}==========================================${NC}"
