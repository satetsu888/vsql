-- Test for type conversion and numeric comparison bugs
-- Bug: All comparisons are done as strings, causing incorrect results

-- Create test table
CREATE TABLE numeric_test (
    id INTEGER,
    price DECIMAL,
    name TEXT,
    quantity INTEGER
);

-- Insert test data with various numeric values
INSERT INTO numeric_test (id, price, name, quantity) VALUES
    (1, 9.99, 'Item A', 100),
    (2, 10.01, 'Item B', 20),
    (10, 2.50, 'Item C', 5),
    (100, 99.99, 'Item D', 1000),
    (20, 5.00, 'Item E', 200);

-- Test 1: Numeric comparison bug - should return 3 rows but may return wrong results
-- Expected: id=10,100,20 (values > 5)
-- Bug result: id=1,10,100,20,2 (string comparison "1">"5", "2"<"5")
SELECT id, name FROM numeric_test WHERE id > 5 ORDER BY id;

-- Test 2: Between operator with numbers
-- Expected: 2 rows (id=10,20)
-- Bug result: may include id=100 due to string comparison
SELECT id, name FROM numeric_test WHERE id BETWEEN 10 AND 50;

-- Test 3: Price comparison with decimals
-- Expected: 2 rows (price > 10.00)
-- Bug result: incorrect due to string comparison
SELECT id, price FROM numeric_test WHERE price > 10.00;

-- Test 4: Multiple numeric comparisons
-- Expected: 1 row (id=20)
-- Bug result: may be incorrect
SELECT * FROM numeric_test WHERE id > 10 AND quantity < 300;

-- Test 5: Comparing numeric columns
-- Expected: 3 rows where id < quantity
-- Bug result: string comparison of columns
SELECT id, quantity FROM numeric_test WHERE id < quantity;

-- Test 6: Arithmetic in WHERE clause (if supported)
-- This may fail completely or give wrong results
SELECT id, price FROM numeric_test WHERE price * 2 > 15;

-- Test 7: IN clause with numbers
-- Expected: 3 rows
-- Bug result: may fail to match due to string comparison
SELECT id, name FROM numeric_test WHERE id IN (1, 10, 100);

-- Test 8: Sorting by numeric column
-- Expected: sorted numerically (1,2,10,20,100)
-- Bug result: sorted as strings (1,10,100,2,20)
SELECT id FROM numeric_test ORDER BY id;

-- Test 9: MAX/MIN aggregates on numeric columns
-- Expected: MAX(id)=100, MIN(id)=1
-- Bug result: MAX may be "99" or "9" due to string comparison
SELECT MAX(id) as max_id, MIN(id) as min_id FROM numeric_test;

-- Test 10: COUNT with numeric condition
-- Expected: 3 rows
-- Bug result: incorrect count
SELECT COUNT(*) as count FROM numeric_test WHERE quantity >= 100;

-- Cleanup
DROP TABLE numeric_test;