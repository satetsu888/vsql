-- Test 13: Self-join
-- Testing LEFT JOIN on the same table

-- Setup
CREATE TABLE categories (id int, name text, parent_id int);

INSERT INTO categories VALUES 
  (1, 'Electronics', NULL),
  (2, 'Computers', 1),
  (3, 'Phones', 1),
  (4, 'Accessories', 1);

-- Test Query
SELECT 
  c1.name as category,
  c2.name as parent_category
FROM categories c1
LEFT JOIN categories c2 ON c1.parent_id = c2.id;

-- Cleanup
DROP TABLE categories;