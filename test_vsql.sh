#!/bin/bash

echo "Starting VSQL server..."
./vsql &
SERVER_PID=$!

sleep 2

echo "Testing connection with psql..."
psql -h localhost -p 5432 -U test -d test <<EOF
CREATE TABLE users;
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
SELECT * FROM users;
SELECT name, email FROM users WHERE id = 1;
SELECT * FROM users WHERE age > 25;
\q
EOF

echo "Stopping VSQL server..."
kill $SERVER_PID

echo "Test completed!"