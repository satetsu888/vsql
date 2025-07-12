-- Test: Multiple correlation conditions
-- Expected: 3 rows (Charlie, David, Eve - employees with salary > department budget / 5)

-- Setup
CREATE TABLE employees (id int, name text, department_id int, salary int);
CREATE TABLE departments (id int, name text, budget int);

INSERT INTO employees VALUES
  (1, 'Alice', 1, 100000),
  (2, 'Bob', 1, 90000),
  (3, 'Charlie', 2, 85000),
  (4, 'David', 2, 95000),
  (5, 'Eve', 3, 110000);

INSERT INTO departments VALUES
  (1, 'Engineering', 500000),
  (2, 'Sales', 300000),
  (3, 'Marketing', 200000);

-- Test query
SELECT name FROM employees e
WHERE EXISTS (
  SELECT 1 FROM departments d 
  WHERE d.id = e.department_id 
  AND e.salary > d.budget / 5
)
ORDER BY name;

-- Cleanup  
DROP TABLE departments;
DROP TABLE employees;