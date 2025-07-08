-- Test 8: CTE (WITH clause) - Complex query
-- Expected: 3 rows with city statistics
-- Status: FAILING
-- FAILING: CTEs (WITH clause) not implemented
-- Test: WITH city_stats AS (...) SELECT * FROM city_stats WHERE total_revenue > 1000 OR user_count > 2 ORDER BY total_revenue DESC

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT,
    age INTEGER,
    city TEXT
);

CREATE TABLE orders (
    id INTEGER,
    user_id INTEGER,
    product TEXT,
    amount INTEGER
);

-- Insert test data
INSERT INTO users (id, name, age, city) VALUES
    (1, 'Alice', 30, 'Tokyo'),
    (2, 'Bob', 25, 'Osaka'),
    (3, 'Charlie', 35, 'Tokyo'),
    (4, 'Dave', 45, 'Kyoto'),
    (5, 'Eve', 28, 'Osaka');

INSERT INTO orders (id, user_id, product, amount) VALUES
    (1, 1, 'Laptop', 1500),
    (2, 1, 'Mouse', 50),
    (3, 2, 'Keyboard', 100),
    (4, 3, 'Monitor', 800),
    (5, 3, 'Laptop', 1600),
    (6, 4, 'Mouse', 45);

-- Test query
WITH city_stats AS (
  SELECT u.city, 
         COUNT(DISTINCT u.id) as user_count,
         COALESCE(SUM(o.amount), 0) as total_revenue,
         COALESCE(AVG(o.amount), 0) as avg_order_value
  FROM users u
  LEFT JOIN orders o ON u.id = o.user_id
  GROUP BY u.city
)
SELECT * FROM city_stats
WHERE total_revenue > 1000 OR user_count > 2
ORDER BY total_revenue DESC;

-- Cleanup
DROP TABLE orders;
DROP TABLE users;