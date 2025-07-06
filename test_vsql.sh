#!/bin/bash

# Simple test script for VSQL

# Cleanup function
cleanup() {
    if [ -n "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Stopping VSQL server..."
        kill $SERVER_PID 2>/dev/null
    fi
}

# Set up trap to cleanup on exit
trap cleanup EXIT INT TERM

echo "Starting VSQL server..."
./vsql 2>/dev/null &
SERVER_PID=$!

echo "Waiting for server to start..."
sleep 2

echo "Testing connection with psql..."
psql -h localhost -p 5432 -U test -d test <<EOF
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
SELECT * FROM users;
SELECT name, email FROM users WHERE id = 1;
SELECT * FROM users WHERE age > 25;
\q
EOF

echo "Test completed!"