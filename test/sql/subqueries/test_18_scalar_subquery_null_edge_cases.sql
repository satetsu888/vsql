-- Test 4: Scalar subqueries with NULL values and edge cases
-- Expected: 4 rows with NULL and edge case handling
-- Test: Scalar subqueries that return NULL or empty results

-- Create test tables
CREATE TABLE categories (
    id INTEGER,
    name TEXT,
    parent_id INTEGER
);

CREATE TABLE items (
    id INTEGER,
    name TEXT,
    category_id INTEGER,
    price INTEGER,
    quantity INTEGER
);

-- Insert test data
INSERT INTO categories (id, name, parent_id) VALUES
    (1, 'Electronics', NULL),
    (2, 'Books', NULL),
    (3, 'Clothing', NULL),
    (4, 'Empty Category', NULL); -- Category with no items

INSERT INTO items (id, name, category_id, price, quantity) VALUES
    (1, 'Laptop', 1, 1000, 5),
    (2, 'Phone', 1, 800, 10),
    (3, 'Novel', 2, 20, NULL), -- NULL quantity
    (4, 'Textbook', 2, 50, 15),
    (5, 'Shirt', 3, 30, NULL), -- NULL quantity
    (6, 'Pants', 3, 40, NULL), -- NULL quantity
    (7, 'Gadget', 1, NULL, 3); -- NULL price

-- Test scalar subqueries with NULL handling
SELECT 
    c.name as category,
    (SELECT COUNT(*) FROM items i WHERE i.category_id = c.id) as item_count,
    (SELECT SUM(i.price) FROM items i WHERE i.category_id = c.id) as total_price,
    (SELECT SUM(i.quantity) FROM items i WHERE i.category_id = c.id) as total_quantity,
    (SELECT AVG(i.price) FROM items i WHERE i.category_id = c.id) as avg_price,
    (SELECT MAX(i.price) FROM items i WHERE i.category_id = c.id AND i.quantity IS NOT NULL) as max_price_with_qty
FROM categories c
ORDER BY c.id;

-- Cleanup
DROP TABLE items;
DROP TABLE categories;