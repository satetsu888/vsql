-- Test: Basic non-correlated EXISTS (should work)
-- Expected: 5 rows (all users) because departments table has rows

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
WHERE EXISTS (SELECT 1 FROM departments)
ORDER BY name;

-- Cleanup
DROP TABLE departments;
DROP TABLE users;