#!/bin/bash

# Memory Stress Test for VSQL
# Tests memory usage with large data volumes and checks for memory leaks

echo "=== VSQL Memory Stress Test ==="
echo "This test will insert large amounts of data and monitor memory usage"
echo ""

# Start VSQL server
echo "Starting VSQL server..."
./vsql &
SERVER_PID=$!

# Wait for server to start
sleep 2

# Function to get memory usage of the VSQL process
get_memory_usage() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        ps -o rss= -p $SERVER_PID | awk '{print $1/1024 " MB"}'
    else
        # Linux
        ps -o rss= -p $SERVER_PID | awk '{print $1/1024 " MB"}'
    fi
}

# Function to run SQL and suppress output
run_sql() {
    psql -h localhost -p 5432 -U test -d test -t -q -c "$1" 2>/dev/null
}

echo "Initial memory usage: $(get_memory_usage)"
echo ""

# Test 1: Large batch inserts
echo "Test 1: Inserting 100,000 rows in batches..."
run_sql "CREATE TABLE large_table (id int, data text, value decimal, created_at timestamp);"

# Insert data in batches of 1000
for i in {1..100}; do
    echo -ne "\rProgress: $i/100 batches"
    
    # Build a large INSERT statement with 1000 rows
    insert_stmt="INSERT INTO large_table VALUES "
    for j in {1..1000}; do
        row_id=$((($i - 1) * 1000 + $j))
        if [ $j -eq 1 ]; then
            insert_stmt+="($row_id, 'data_${row_id}_$(head -c 100 < /dev/urandom | base64 | tr -d '\n')', $((RANDOM % 10000)).99, NOW())"
        else
            insert_stmt+=", ($row_id, 'data_${row_id}_$(head -c 100 < /dev/urandom | base64 | tr -d '\n')', $((RANDOM % 10000)).99, NOW())"
        fi
    done
    insert_stmt+=";"
    
    run_sql "$insert_stmt"
done

echo -e "\nMemory usage after 100k inserts: $(get_memory_usage)"

# Test 2: Complex queries on large dataset
echo -e "\nTest 2: Running complex queries on large dataset..."
for i in {1..10}; do
    echo -ne "\rRunning query $i/10"
    run_sql "SELECT id, data, value FROM large_table WHERE value > 5000 ORDER BY value DESC LIMIT 1000;"
    run_sql "SELECT COUNT(*), AVG(value), MAX(value), MIN(value) FROM large_table WHERE id BETWEEN $((i * 1000)) AND $(((i + 1) * 1000));"
done

echo -e "\nMemory usage after queries: $(get_memory_usage)"

# Test 3: Multiple table operations
echo -e "\nTest 3: Creating multiple tables with data..."
for i in {1..20}; do
    echo -ne "\rCreating table $i/20"
    run_sql "CREATE TABLE table_$i (id int, name text, data text);"
    
    # Insert 5000 rows per table
    insert_stmt="INSERT INTO table_$i VALUES "
    for j in {1..5000}; do
        if [ $j -eq 1 ]; then
            insert_stmt+="($j, 'name_$j', 'data_$(head -c 50 < /dev/urandom | base64 | tr -d '\n')')"
        else
            insert_stmt+=", ($j, 'name_$j', 'data_$(head -c 50 < /dev/urandom | base64 | tr -d '\n')')"
        fi
    done
    insert_stmt+=";"
    run_sql "$insert_stmt"
done

echo -e "\nMemory usage after creating 20 tables: $(get_memory_usage)"

# Test 4: Stress test with updates and deletes
echo -e "\nTest 4: Performing updates and deletes..."
for i in {1..50}; do
    echo -ne "\rUpdate/Delete cycle $i/50"
    # Update random rows
    run_sql "UPDATE large_table SET data = 'updated_$i' WHERE id % 100 = $i;"
    # Delete some rows
    run_sql "DELETE FROM large_table WHERE id % 200 = $i;"
done

echo -e "\nMemory usage after updates/deletes: $(get_memory_usage)"

# Test 5: Join operations on large tables
echo -e "\nTest 5: Testing memory usage with JOINs..."
run_sql "CREATE TABLE users (id int, name text, email text);"
run_sql "CREATE TABLE orders (id int, user_id int, amount decimal);"

# Insert test data
for i in {1..10000}; do
    if [ $((i % 1000)) -eq 0 ]; then
        echo -ne "\rInserting users/orders: $i/10000"
    fi
    run_sql "INSERT INTO users VALUES ($i, 'user_$i', 'user$i@example.com');"
    # Each user has 1-5 orders
    num_orders=$((RANDOM % 5 + 1))
    for j in {1..$num_orders}; do
        order_id=$(($i * 10 + $j))
        run_sql "INSERT INTO orders VALUES ($order_id, $i, $((RANDOM % 1000)).99);"
    done
done

echo -e "\nRunning JOIN queries..."
for i in {1..10}; do
    echo -ne "\rJOIN query $i/10"
    run_sql "SELECT u.name, COUNT(o.id) as order_count, SUM(o.amount) as total 
             FROM users u 
             LEFT JOIN orders o ON u.id = o.user_id 
             GROUP BY u.id, u.name 
             HAVING COUNT(o.id) > 2
             LIMIT 100;"
done

echo -e "\nMemory usage after JOINs: $(get_memory_usage)"

# Test 6: Table drops and memory release
echo -e "\nTest 6: Dropping tables to test memory release..."
for i in {1..20}; do
    echo -ne "\rDropping table $i/20"
    run_sql "DROP TABLE table_$i;"
done

# Give some time for garbage collection (if any)
sleep 5

echo -e "\nMemory usage after dropping 20 tables: $(get_memory_usage)"

# Test 7: Extreme stress - create and drop repeatedly
echo -e "\nTest 7: Create/drop cycle stress test..."
initial_mem=$(get_memory_usage | cut -d' ' -f1)
for cycle in {1..20}; do
    echo -ne "\rCycle $cycle/20"
    
    # Create table with data
    run_sql "CREATE TABLE stress_test (id int, data text);"
    insert_stmt="INSERT INTO stress_test VALUES "
    for j in {1..1000}; do
        if [ $j -eq 1 ]; then
            insert_stmt+="($j, 'data_$(head -c 200 < /dev/urandom | base64 | tr -d '\n')')"
        else
            insert_stmt+=", ($j, 'data_$(head -c 200 < /dev/urandom | base64 | tr -d '\n')')"
        fi
    done
    insert_stmt+=";"
    run_sql "$insert_stmt"
    
    # Drop table
    run_sql "DROP TABLE stress_test;"
done

final_mem=$(get_memory_usage | cut -d' ' -f1)
echo -e "\nMemory usage after create/drop cycles: $(get_memory_usage)"

# Calculate memory growth
mem_growth=$(echo "$final_mem - $initial_mem" | bc)
echo -e "\nMemory growth during stress test: $mem_growth MB"

# Final summary
echo -e "\n=== Test Summary ==="
echo "Final memory usage: $(get_memory_usage)"
echo "Server PID: $SERVER_PID"

# Check if server is still responsive
if run_sql "SELECT 1;" > /dev/null 2>&1; then
    echo "Server is still responsive ✓"
else
    echo "Server is not responsive ✗"
fi

# Cleanup
echo -e "\nCleaning up..."
run_sql "DROP TABLE IF EXISTS large_table;"
run_sql "DROP TABLE IF EXISTS users;"
run_sql "DROP TABLE IF EXISTS orders;"

echo "Stopping VSQL server..."
kill $SERVER_PID 2>/dev/null

echo -e "\nMemory stress test completed!"