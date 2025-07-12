-- Test: Correlated EXISTS with additional conditions
-- Expected: 2 rows (Alice, David - salary > avg in their department)

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
  SELECT 1 FROM employees e2 
  WHERE e2.department_id = e.department_id 
  AND e.salary > (
    SELECT AVG(salary) FROM employees e3 
    WHERE e3.department_id = e.department_id
  )
)
ORDER BY name;

-- Cleanup  
DROP TABLE departments;
DROP TABLE employees;