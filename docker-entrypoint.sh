#!/bin/bash
set -e

# Default seed directory
SEED_DIR="${SEED_DIR:-/seed}"

# Array to hold all -f arguments for seed files
SEED_FILES=()

# Check if seed directory exists and contains SQL files
if [ -d "$SEED_DIR" ] && [ -n "$(ls -A $SEED_DIR/*.sql 2>/dev/null)" ]; then
    echo "Found SQL files in seed directory: $SEED_DIR"
    
    # Sort files to ensure consistent order
    for file in $(ls $SEED_DIR/*.sql | sort); do
        echo "  Adding seed file: $file"
        SEED_FILES+=("-f" "$file")
    done
fi

# Check if we need to add -q flag (quit after execution)
# If user provides no arguments and we have seed files, start as server after seeding
QUIT_FLAG=""
if [ $# -eq 0 ] && [ ${#SEED_FILES[@]} -eq 0 ]; then
    # No arguments and no seed files - just start server
    exec /app/vsql
elif [ $# -gt 0 ]; then
    # User provided arguments - check if -q is already present
    for arg in "$@"; do
        if [ "$arg" = "-q" ]; then
            QUIT_FLAG="exists"
            break
        fi
    done
fi

# Execute vsql with seed files and any user-provided arguments
if [ ${#SEED_FILES[@]} -gt 0 ]; then
    echo "Executing VSQL with seed files..."
    exec /app/vsql "${SEED_FILES[@]}" "$@"
else
    # No seed files, just pass through arguments
    exec /app/vsql "$@"
fi