#!/bin/bash
set -euo pipefail

# generate-all.sh
# Batch generate documentation for all profile × template combinations
# against the test fixture app.
#
# Usage: ./test/scripts/generate-all.sh
#
# Env vars:
#   BINARY      - goscribe binary path (default: ./goscribe)
#   TEST_APP    - test fixture app path (default: test/app)
#   OUTPUT_BASE - output base directory (default: test/doc)

BINARY="${BINARY:-./goscribe}"
TEST_APP="${TEST_APP:-test/app}"
OUTPUT_BASE="${OUTPUT_BASE:-test/doc}"

PROFILES=(
    software-documenter
    technical-writer
    github-readme-expert
    github-wiki-expert
    api-reference
    developer-onboarding
    operations-runbook
    release-notes
    architecture-overview
    contributing-guide
    package-reference
)

TEMPLATES=(elegant technical futuristic minimal friendly)

SUCCESS=0
FAILED=0
TOTAL=0

echo "=== GoScribe Generate All ==="
echo "Binary:     $BINARY"
echo "Test app:   $TEST_APP"
echo "Output:     $OUTPUT_BASE"
echo "Profiles:   ${#PROFILES[@]}"
echo "Templates:  ${#TEMPLATES[@]}"
echo ""

for PROFILE in "${PROFILES[@]}"; do
    # Run without template
    echo "--- $PROFILE / none ---"
    if "$BINARY" generate "$TEST_APP" --profile "$PROFILE" -o "$OUTPUT_BASE/$PROFILE/none" 2>&1; then
        SUCCESS=$((SUCCESS + 1))
    else
        FAILED=$((FAILED + 1))
    fi
    TOTAL=$((TOTAL + 1))

    for TEMPLATE in "${TEMPLATES[@]}"; do
        echo "--- $PROFILE / $TEMPLATE ---"
        if "$BINARY" generate "$TEST_APP" --profile "$PROFILE" --template "$TEMPLATE" -o "$OUTPUT_BASE/$PROFILE/$TEMPLATE" 2>&1; then
            SUCCESS=$((SUCCESS + 1))
        else
            FAILED=$((FAILED + 1))
        fi
        TOTAL=$((TOTAL + 1))
    done
done

echo ""
echo "=========================================="
echo " Generated: $SUCCESS combinations. Failed: $FAILED."
echo " Total:     $TOTAL"
echo "=========================================="

if [ "$FAILED" -gt 0 ]; then
    exit 1
fi
exit 0
