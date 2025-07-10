-- Test 3: Scalar subqueries with arithmetic operations
-- Expected: 3 rows with calculated metrics
-- Test: Scalar subqueries used in arithmetic expressions

-- Create test tables
CREATE TABLE stores (
    id INTEGER,
    name TEXT,
    region TEXT
);

CREATE TABLE sales (
    id INTEGER,
    store_id INTEGER,
    product TEXT,
    amount INTEGER,
    sale_date TEXT
);

CREATE TABLE targets (
    store_id INTEGER,
    target_amount INTEGER
);

-- Insert test data
INSERT INTO stores (id, name, region) VALUES
    (1, 'Store A', 'North'),
    (2, 'Store B', 'South'),
    (3, 'Store C', 'East');

INSERT INTO sales (id, store_id, product, amount, sale_date) VALUES
    (1, 1, 'Product X', 1000, '2024-01-01'),
    (2, 1, 'Product Y', 1500, '2024-01-02'),
    (3, 1, 'Product Z', 2000, '2024-01-03'),
    (4, 2, 'Product X', 800, '2024-01-01'),
    (5, 2, 'Product Y', 1200, '2024-01-02'),
    (6, 3, 'Product X', 500, '2024-01-01');

INSERT INTO targets (store_id, target_amount) VALUES
    (1, 5000),
    (2, 3000),
    (3, 2000);

-- Test scalar subqueries in arithmetic expressions
SELECT 
    s.name as store_name,
    (SELECT SUM(amount) FROM sales WHERE store_id = s.id) as total_sales,
    (SELECT target_amount FROM targets WHERE store_id = s.id) as target,
    (SELECT SUM(amount) FROM sales WHERE store_id = s.id) - 
        (SELECT target_amount FROM targets WHERE store_id = s.id) as variance,
    CASE 
        WHEN (SELECT SUM(amount) FROM sales WHERE store_id = s.id) >= 
             (SELECT target_amount FROM targets WHERE store_id = s.id) 
        THEN 'Met'
        ELSE 'Not Met'
    END as target_status
FROM stores s
ORDER BY s.id;

-- Cleanup
DROP TABLE targets;
DROP TABLE sales;
DROP TABLE stores;