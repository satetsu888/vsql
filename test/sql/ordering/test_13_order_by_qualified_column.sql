-- Test: ORDER BY with qualified column names (table.column)
-- Expected: 3 rows

CREATE TABLE products (id int, name text, price int);
CREATE TABLE categories (id int, name text);

INSERT INTO products VALUES (1, 'Laptop', 1000), (2, 'Mouse', 20), (3, 'Keyboard', 50);
INSERT INTO categories VALUES (1, 'Electronics'), (2, 'Accessories');

-- Use qualified column names in ORDER BY
SELECT p.id, p.name, p.price, c.name AS category
FROM products p
CROSS JOIN categories c
WHERE c.id = 1
ORDER BY p.price DESC, p.id ASC;

DROP TABLE products;
DROP TABLE categories;