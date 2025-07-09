-- Test 10: Aggregate functions with WHERE clause filtering
-- Expected: Aggregates should apply only to filtered rows

-- Create test table
CREATE TABLE sales (
    id INTEGER,
    product TEXT,
    category TEXT,
    amount DECIMAL,
    quantity INTEGER
);

-- Insert test data
INSERT INTO sales (id, product, category, amount, quantity) VALUES
    (1, 'Laptop', 'Electronics', 1200.00, 1),
    (2, 'Mouse', 'Electronics', 25.00, 5),
    (3, 'Desk', 'Furniture', 300.00, 2),
    (4, 'Chair', 'Furniture', 150.00, 4),
    (5, 'Monitor', 'Electronics', 400.00, 2),
    (6, 'Keyboard', 'Electronics', 75.00, 3);

-- Test query 1: Aggregates with WHERE clause
SELECT 
    COUNT(*) as electronics_count,
    SUM(amount) as total_amount,
    AVG(amount) as avg_amount,
    MAX(quantity) as max_quantity,
    MIN(quantity) as min_quantity
FROM sales
WHERE category = 'Electronics';

-- Test query 2: Aggregates with complex WHERE clause
SELECT 
    COUNT(*) as expensive_items,
    SUM(amount * quantity) as total_revenue
FROM sales
WHERE amount > 100 AND quantity >= 2;

-- Cleanup
DROP TABLE sales;