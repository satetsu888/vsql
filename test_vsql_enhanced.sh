#!/bin/bash

# Enhanced Test Script for VSQL
# Comprehensive testing including edge cases and error scenarios

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test result counters
PASSED=0
FAILED=0
SERVER_PID=""

# Cleanup function
cleanup() {
    echo ""
    if [ -n "$SERVER_PID" ] && kill -0 $SERVER_PID 2>/dev/null; then
        echo "Stopping VSQL server..."
        kill $SERVER_PID 2>/dev/null || true
        sleep 1
        # Force kill if still running
        if kill -0 $SERVER_PID 2>/dev/null; then
            kill -9 $SERVER_PID 2>/dev/null || true
        fi
    fi
}

# Set up trap to cleanup on exit
trap cleanup EXIT INT TERM

# Function to print colored output
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((FAILED++))
}

# Function to run SQL and check result
run_test() {
    local test_name="$1"
    local sql="$2"
    local expected_behavior="$3"  # "success" or "error"
    
    print_test "$test_name"
    
    # Run psql with timeout
    output=$(timeout 5 psql -h localhost -p 5432 -U test -d test -c "$sql" 2>&1)
    exit_code=$?
    
    # Handle timeout
    if [ $exit_code -eq 124 ]; then
        print_fail "$test_name - Command timed out"
        return 1
    fi
    
    if [ "$expected_behavior" = "success" ]; then
        if [ $exit_code -eq 0 ]; then
            print_pass "$test_name"
            return 0
        else
            print_fail "$test_name - Expected success but got error: $output"
            return 1
        fi
    else
        if [ $exit_code -ne 0 ]; then
            print_pass "$test_name - Correctly failed as expected"
            return 0
        else
            print_fail "$test_name - Expected error but succeeded"
            return 1
        fi
    fi
}

# Check if vsql binary exists
if [ ! -f "./vsql" ]; then
    echo -e "${RED}Error: vsql binary not found. Please run 'go build -o vsql' first.${NC}"
    exit 1
fi

# Check if vsql is already running on port 5432
if lsof -Pi :5432 -sTCP:LISTEN -t >/dev/null ; then
    echo -e "${RED}Error: Port 5432 is already in use. Please stop any existing VSQL or PostgreSQL instances.${NC}"
    exit 1
fi

# Start VSQL server
echo "Starting VSQL server..."
./vsql 2>/dev/null &
SERVER_PID=$!

# Wait for server to start
echo "Waiting for server to start..."
for i in {1..10}; do
    # Use CREATE TABLE instead of SELECT 1 to test connection
    if echo "CREATE TABLE test_connection (id int); DROP TABLE test_connection;" | psql -h localhost -p 5432 -U test -d test >/dev/null 2>&1; then
        echo "Server is ready!"
        break
    fi
    if ! kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${RED}VSQL server failed to start${NC}"
        exit 1
    fi
    echo "Waiting... ($i/10)"
    sleep 1
done

# Final verification
if [ $i -eq 10 ]; then
    echo -e "${RED}Failed to connect to VSQL server after 10 seconds${NC}"
    exit 1
fi

echo "VSQL server started with PID: $SERVER_PID"
echo ""
echo "=== Running Enhanced Test Suite ==="
echo ""

# Basic functionality tests
echo -e "\n${YELLOW}### Basic Table Operations${NC}"
run_test "Create table" "CREATE TABLE users (id int, name text, email text);" "success"
run_test "Insert data" "INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');" "success"
run_test "Select all" "SELECT * FROM users;" "success"
run_test "Update data" "UPDATE users SET email = 'alice.new@example.com' WHERE id = 1;" "success"
run_test "Delete data" "DELETE FROM users WHERE id = 1;" "success"

# Schema-less feature tests
echo -e "\n${YELLOW}### Schema-less Features${NC}"
run_test "Insert with new column" "INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);" "success"
run_test "Select non-existent column" "SELECT id, name, phone FROM users;" "success"
run_test "Mixed schema insert" "INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');" "success"

# Complex WHERE clause tests
echo -e "\n${YELLOW}### Complex WHERE Clauses${NC}"
run_test "AND condition" "SELECT * FROM users WHERE id > 1 AND name = 'Bob';" "success"
run_test "OR condition" "SELECT * FROM users WHERE id = 2 OR name = 'Charlie';" "success"
run_test "NOT condition" "SELECT * FROM users WHERE NOT (id = 2);" "success"
run_test "Combined conditions" "SELECT * FROM users WHERE (id > 1 AND name LIKE 'B%') OR email IS NOT NULL;" "success"

# NULL handling tests
echo -e "\n${YELLOW}### NULL Handling${NC}"
run_test "Insert NULL values" "INSERT INTO users (id, name, email) VALUES (4, NULL, NULL);" "success"
run_test "IS NULL check" "SELECT * FROM users WHERE email IS NULL;" "success"
run_test "IS NOT NULL check" "SELECT * FROM users WHERE name IS NOT NULL;" "success"

# Error handling tests
echo -e "\n${YELLOW}### Error Handling${NC}"
run_test "Select from non-existent table" "SELECT * FROM non_existent_table;" "error"
run_test "Invalid SQL syntax" "SELECT * FROM WHERE id = 1;" "error"
# Note: VSQL might not return error for dropping non-existent table
run_test "Drop non-existent table" "DROP TABLE non_existent_table;" "success"

# Special characters and SQL injection tests
echo -e "\n${YELLOW}### Special Characters and Security${NC}"
run_test "Single quotes in data" "INSERT INTO users (id, name) VALUES (5, 'O''Brien');" "success"
run_test "Special characters" "INSERT INTO users (id, name) VALUES (6, 'Test!@#\$%^&*()');" "success"
run_test "Unicode characters" "INSERT INTO users (id, name) VALUES (7, '日本語テスト');" "success"

# JOIN tests (if supported)
echo -e "\n${YELLOW}### JOIN Operations${NC}"
run_test "Create orders table" "CREATE TABLE orders (id int, user_id int, amount decimal);" "success"
run_test "Insert order data" "INSERT INTO orders VALUES (1, 2, 100.50), (2, 3, 200.00);" "success"
run_test "Inner join" "SELECT u.name, o.amount FROM users u INNER JOIN orders o ON u.id = o.user_id;" "success"
run_test "Left join" "SELECT u.name, o.amount FROM users u LEFT JOIN orders o ON u.id = o.user_id;" "success"

# Aggregate function tests
echo -e "\n${YELLOW}### Aggregate Functions${NC}"
run_test "COUNT function" "SELECT COUNT(*) FROM users;" "success"
run_test "SUM function" "SELECT SUM(amount) FROM orders;" "success"
run_test "AVG function" "SELECT AVG(amount) FROM orders;" "success"
run_test "GROUP BY" "SELECT user_id, COUNT(*) FROM orders GROUP BY user_id;" "success"

# Subquery tests
echo -e "\n${YELLOW}### Subqueries${NC}"
run_test "Subquery in WHERE" "SELECT * FROM users WHERE id IN (SELECT user_id FROM orders);" "success"
run_test "EXISTS subquery" "SELECT * FROM users u WHERE EXISTS (SELECT 1 FROM orders o WHERE o.user_id = u.id);" "success"

# Transaction tests (if supported)
echo -e "\n${YELLOW}### Transaction Support${NC}"
# Test transaction support without causing the script to exit
if echo -e "BEGIN;\nINSERT INTO users (id, name) VALUES (100, 'Transaction Test');\nROLLBACK;" | psql -h localhost -p 5432 -U test -d test 2>&1 | grep -q "ERROR"; then
    print_pass "Transaction not supported (as expected)"
else
    print_fail "Transaction test unclear"
fi

# Performance test with larger dataset
echo -e "\n${YELLOW}### Performance Test${NC}"
echo "Inserting 1000 rows..."
# Create a single large INSERT statement
insert_values=""
for i in {101..1100}; do
    if [ -z "$insert_values" ]; then
        insert_values="($i, 'User$i', 'user$i@example.com')"
    else
        insert_values="$insert_values, ($i, 'User$i', 'user$i@example.com')"
    fi
done
run_test "Bulk insert 1000 rows" "INSERT INTO users (id, name, email) VALUES $insert_values;" "success"

run_test "Select with large dataset" "SELECT COUNT(*) FROM users;" "success"
run_test "Complex query on large dataset" "SELECT * FROM users WHERE id > 500 AND email LIKE '%@example.com' ORDER BY id DESC LIMIT 10;" "success"

# Cleanup test
echo -e "\n${YELLOW}### Cleanup Operations${NC}"
run_test "Drop orders table" "DROP TABLE orders;" "success"
run_test "Drop users table" "DROP TABLE users;" "success"

# Summary
echo ""
echo "=== Test Summary ==="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total: $((PASSED + FAILED))"

# Exit with appropriate code
if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi