-- Test: Correlated EXISTS subqueries limitations
-- Status: Expected to FAIL - correlated subqueries not fully supported
-- These tests document what doesn't work yet with EXISTS

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

-- Test 1: Simple correlated EXISTS
-- Expected to FAIL: References outer table 'e' in subquery
-- Should return: Alice, Bob, Charlie, David, Eve (all have matching departments)
SELECT name FROM employees e
WHERE EXISTS (SELECT 1 FROM departments d WHERE d.id = e.department_id)
ORDER BY name;

-- Test 2: Correlated EXISTS with additional conditions
-- Expected to FAIL: References outer table 'e' in subquery  
-- Should return: Alice, Eve (salary > avg in their department)
SELECT name FROM employees e
WHERE EXISTS (
  SELECT 1 FROM employees e2 
  WHERE e2.department_id = e.department_id 
  AND e.salary > (
    SELECT AVG(salary) FROM employees e3 
    WHERE e3.department_id = e.department_id
  )
)
ORDER BY name;

-- Test 3: Correlated NOT EXISTS
-- Expected to FAIL: References outer table 'e' in subquery
-- Should return: empty (all employees have departments)
SELECT name FROM employees e
WHERE NOT EXISTS (SELECT 1 FROM departments d WHERE d.id = e.department_id);

-- Test 4: Multiple correlation conditions
-- Expected to FAIL: Multiple references to outer table
-- Should return: Alice, Eve (employees with salary > department budget / 5)
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