#\!/bin/bash

# Start VSQL in the background
./vsql &
VSQL_PID=$!

# Wait for server to start
sleep 2

# Create test table and data
psql -h localhost -p 5432 -U test -d test -c 'CREATE TABLE "Todo" ("id" integer, "title" text, "completed" boolean, "createdAt" timestamp, "updatedAt" timestamp);'
psql -h localhost -p 5432 -U test -d test -c "INSERT INTO \"Todo\" (\"id\", \"title\", \"completed\", \"createdAt\", \"updatedAt\") VALUES (1, 'Test Todo', false, '2024-01-01 10:00:00', '2024-01-01 10:00:00');"

# Test a parameterized query similar to what the ORM uses
psql -h localhost -p 5432 -U test -d test -c 'SELECT "id", "title", "completed", "createdAt", "updatedAt" FROM "Todo" WHERE 1=1 ORDER BY "createdAt" DESC OFFSET 0 LIMIT 10;' -v ON_ERROR_STOP=1

# Clean up
kill $VSQL_PID
wait $VSQL_PID 2>/dev/null

echo "Test completed"
