-- Test: Correlated NOT EXISTS
-- Expected: no rows (all employees have departments)
-- Status: FAILING - References outer table 'e' in subquery

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
WHERE NOT EXISTS (SELECT 1 FROM departments d WHERE d.id = e.department_id);

-- Cleanup  
DROP TABLE departments;
DROP TABLE employees;