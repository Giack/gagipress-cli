#!/bin/bash
set -e

echo "=================================================="
echo "E2E Test Verification Script"
echo "=================================================="
echo ""

# Check if Supabase credentials are set
if [ -z "$SUPABASE_URL" ]; then
    echo "❌ ERROR: SUPABASE_URL environment variable not set"
    echo ""
    echo "Please set your Supabase credentials:"
    echo "  export SUPABASE_URL='https://your-project.supabase.co'"
    echo "  export SUPABASE_KEY='your-anon-key'"
    echo "  export SUPABASE_SERVICE_KEY='your-service-key'  # Optional but recommended"
    echo ""
    exit 1
fi

if [ -z "$SUPABASE_KEY" ]; then
    echo "❌ ERROR: SUPABASE_KEY environment variable not set"
    exit 1
fi

echo "✅ Supabase credentials found"
echo "   URL: $SUPABASE_URL"
echo "   Key: ${SUPABASE_KEY:0:20}..."
if [ -n "$SUPABASE_SERVICE_KEY" ]; then
    echo "   Service Key: ${SUPABASE_SERVICE_KEY:0:20}..."
else
    echo "   ⚠️  Service Key: Not set (using anon key)"
fi
echo ""

echo "=================================================="
echo "Running E2E Tests"
echo "=================================================="
echo ""

# Run the tests
if mise exec -- go test ./test/integration/approve_reject_test.go \
    ./test/integration/fixtures.go \
    ./test/integration/setup_test.go -v; then
    echo ""
    echo "=================================================="
    echo "✅ All E2E Tests Passed!"
    echo "=================================================="
    echo ""
    echo "Fixes verified:"
    echo "  ✅ URL encoding bug fixed (UUID prefix resolution works)"
    echo "  ✅ Schema mismatch fixed (PostType field added)"
    echo "  ✅ Prefer header added (status updates work)"
    echo ""
    echo "Next steps:"
    echo "  1. Test commands manually:"
    echo "     bin/gagipress ideas list"
    echo "     bin/gagipress ideas approve <prefix>"
    echo "     bin/gagipress calendar approve"
    echo ""
    echo "  2. Check fixture cleanup logs for any manual cleanup needed"
    echo ""
else
    echo ""
    echo "=================================================="
    echo "❌ Some Tests Failed"
    echo "=================================================="
    echo ""
    echo "Please review the output above for details."
    echo ""
    exit 1
fi
