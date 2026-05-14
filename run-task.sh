#!/bin/bash

# AIPROXY Migration Task Runner
# Usage: ./run-task.sh <task-id>

set -e

TASK_ID="$1"

if [ -z "$TASK_ID" ]; then
    echo "Usage: ./run-task.sh <task-id>"
    echo ""
    echo "Examples:"
    echo "  ./run-task.sh T0.1"
    echo "  ./run-task.sh T2.5"
    echo "  ./run-task.sh T4.9"
    echo ""
    echo "Available tasks:"
    find tasks -name "*.md" -type f | grep -v README | sort
    exit 1
fi

# Find task file
TASK_FILE=$(find tasks -name "${TASK_ID}-*.md" -type f | head -1)

if [ -z "$TASK_FILE" ]; then
    echo "ERROR: Task ${TASK_ID} not found"
    exit 1
fi

echo "=========================================="
echo "TASK: ${TASK_ID}"
echo "FILE: ${TASK_FILE}"
echo "=========================================="
echo ""

# Generate prompt
cat << 'EOF'
Read and follow strictly:

- EXECUTION_LOOP.md
- BUILD_RULES.md
- AGENT_RULES.md
- PROJECT_CONSTRAINTS.md
- TASK_STATUS.md

Then execute ONLY this task:

EOF

echo "${TASK_FILE}"

cat << 'EOF'

Rules:
- Do NOT continue to another task
- Do NOT modify unrelated files
- Do NOT redesign architecture
- Do NOT fix unrelated build errors
- Read FIRST root build error only
- Preserve 9router behavior parity
- Stop immediately after task completion
- Maximum delegation depth: 1

Required output:
1. Files modified
2. Changes made
3. Verification results
4. TASK_STATUS.md update
5. Any blockers

STOP after task execution.
EOF
