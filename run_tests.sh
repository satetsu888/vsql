#!/bin/bash

# Run all tests for VSQL project

set -e

echo "=== Running VSQL Tests ==="
echo ""

# Run unit tests
echo "1. Running unit tests..."
go test ./... -v

echo ""
echo "2. Running tests with race detector..."
go test ./... -race

echo ""
echo "3. Running test coverage..."
go test ./... -cover

echo ""
echo "=== All tests passed! ==="