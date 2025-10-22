#!/bin/bash

# Script to generate mocks for testing
# Usage: ./scripts/generate-mocks.sh

set -e

GOPATH=$(go env GOPATH)
MOCKGEN="$GOPATH/bin/mockgen"

echo "🔨 Generating mocks..."

# Check if mockgen is installed
if [ ! -f "$MOCKGEN" ]; then
    echo "❌ mockgen not found. Installing..."
    go install go.uber.org/mock/mockgen@latest
fi

# Create mocks directory
mkdir -p internal/service/mocks

# Generate Querier mock
echo "📝 Generating Querier mock..."
$MOCKGEN -source=internal/data/db/querier.go -destination=internal/service/mocks/mock_querier.go -package=mocks

echo "✅ Mocks generated successfully!"
echo ""
echo "Generated files:"
echo "  - internal/service/mocks/mock_querier.go"

