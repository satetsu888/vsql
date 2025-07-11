-- Test: Non-correlated EXISTS subqueries (should work)
-- These tests verify the basic EXISTS structure without correlation
-- Expected: Tests should pass as basic EXISTS structure is implemented

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

-- Test 1: Basic non-correlated EXISTS (should work)
-- Expected: 5 rows (all users) because departments table has rows
SELECT name FROM users
WHERE EXISTS (SELECT 1 FROM departments)
ORDER BY name;

-- Test 2: Non-correlated EXISTS with WHERE clause in subquery
-- Expected: 5 rows (all users) because there are active departments
SELECT name FROM users  
WHERE EXISTS (SELECT 1 FROM departments WHERE active = true)
ORDER BY name;

-- Test 3: Non-correlated NOT EXISTS
-- Expected: 0 rows because departments table has rows
SELECT name FROM users
WHERE NOT EXISTS (SELECT 1 FROM departments);

-- Test 4: Non-correlated NOT EXISTS with impossible condition
-- Expected: 5 rows (all users) because no department named 'NonExistent'
SELECT name FROM users
WHERE NOT EXISTS (SELECT 1 FROM departments WHERE name = 'NonExistent')
ORDER BY name;

-- Test 5: Non-correlated EXISTS returning no rows
-- Expected: 0 rows because no department named 'NonExistent'
SELECT name FROM users
WHERE EXISTS (SELECT 1 FROM departments WHERE name = 'NonExistent');

-- Cleanup
DROP TABLE departments;
DROP TABLE users;