-- Test: Non-correlated EXISTS returning no rows
-- Expected: 0 rows because no department named 'NonExistent'

-- Setup
CREATE TABLE users (id int, name text, department text);
CREATE TABLE departments (name text, active boolean);

INSERT INTO users VALUES
  (1, 'Alice', 'Engineering'),
  (2, 'Bob', 'Sales'),
  (3, 'Charlie', 'Marketing'),
  (4, 'David', 'Engineering'),
  (5, 'Eve', 'HR');

INSERT INTO departments VALUES
  ('Engineering', true),
  ('Sales', true),
  ('Marketing', false),
  ('Finance', true);

-- Test query
SELECT name FROM users
WHERE EXISTS (SELECT 1 FROM departments WHERE name = 'NonExistent');

-- Cleanup
DROP TABLE departments;
DROP TABLE users;