#!/usr/bin/env bash
# proto-gen.sh generates Go code from all .proto files in the proto/ directory.
#
# Output goes to internal/<service>/gen/proto/ so generated code
# lives inside the service that owns it, under internal/ protection.
#
# Run this whenever you change a .proto file:
#   make proto

set -euo pipefail

# Project root is one level up from this script
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Where proto source files live
PROTO_DIR="$ROOT/proto"

echo "Generating proto files from $PROTO_DIR"

# Generate Go code for each service proto
for service in trip driver payment; do
    PROTO_FILE="$PROTO_DIR/$service/$service.proto"

    if [ ! -f "$PROTO_FILE" ]; then
        echo "  skipping $service — $PROTO_FILE not found"
        continue
    fi

    # Output directory — inside the service's internal package
    OUT_DIR="$ROOT/internal/$service/gen/proto"
    mkdir -p "$OUT_DIR"

    echo "  generating $service..."

    protoc \
        --proto_path="$PROTO_DIR" \
        --go_out="$OUT_DIR" \
        --go_opt=paths=source_relative \
        --go-grpc_out="$OUT_DIR" \
        --go-grpc_opt=paths=source_relative \
        "$PROTO_FILE"

    echo "  done → $OUT_DIR"
done

echo "Proto generation complete."