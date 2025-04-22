#!/bin/bash

# Navigate to project root
cd "$(dirname "$0")/.."

# Clean existing generated files
rm -rf api/protos/candlestick

# Ensure Go binary path is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Generate protobuf files with correct paths
protoc --proto_path=api/protos \
    --go_out=. \
    --go-grpc_out=. \
    api/protos/candlestick.proto

# Fix imports if needed
if command -v goimports > /dev/null; then
    find api/protos -name "*.pb.go" -exec goimports -w {} \;
else
    echo "goimports not found, skipping formatting"
fi