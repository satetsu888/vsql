-- Test 5: Scalar subqueries in SELECT - various cases
-- Expected: 5 rows with correct calculated values
-- Test: Multiple scalar subqueries with various aggregate functions

-- Create test tables
CREATE TABLE departments (
    id INTEGER,
    name TEXT,
    budget INTEGER
);

CREATE TABLE employees (
    id INTEGER,
    name TEXT,
    dept_id INTEGER,
    salary INTEGER,
    hire_date TEXT
);

CREATE TABLE projects (
    id INTEGER,
    name TEXT,
    dept_id INTEGER,
    budget INTEGER,
    status TEXT
);

-- Insert test data
INSERT INTO departments (id, name, budget) VALUES
    (1, 'Engineering', 500000),
    (2, 'Sales', 300000),
    (3, 'Marketing', 200000),
    (4, 'HR', 100000),
    (5, 'Support', 150000);

INSERT INTO employees (id, name, dept_id, salary, hire_date) VALUES
    (1, 'John', 1, 120000, '2020-01-15'),
    (2, 'Jane', 1, 130000, '2019-03-20'),
    (3, 'Bob', 2, 90000, '2021-06-01'),
    (4, 'Alice', 2, 95000, '2020-11-10'),
    (5, 'Charlie', 3, 80000, '2022-02-28'),
    (6, 'Eve', 4, 70000, '2021-08-15'),
    (7, 'Dave', 1, 110000, '2023-01-01'),
    (8, 'Frank', 5, 65000, '2022-09-10'),
    (9, 'Grace', 5, 60000, '2023-03-15');

INSERT INTO projects (id, name, dept_id, budget, status) VALUES
    (1, 'Project Alpha', 1, 100000, 'active'),
    (2, 'Project Beta', 1, 150000, 'active'),
    (3, 'Sales Campaign', 2, 50000, 'completed'),
    (4, 'Marketing Push', 3, 80000, 'active'),
    (5, 'Training Program', 4, 30000, 'planned'),
    (6, 'Support System', 5, 40000, 'active'),
    (7, 'Project Gamma', 1, 200000, 'planned');

-- Test multiple scalar subqueries with different aggregates
SELECT 
    d.name as department,
    d.budget,
    (SELECT COUNT(*) FROM employees e WHERE e.dept_id = d.id) as employee_count,
    (SELECT AVG(salary) FROM employees e WHERE e.dept_id = d.id) as avg_salary,
    (SELECT MAX(salary) FROM employees e WHERE e.dept_id = d.id) as max_salary,
    (SELECT MIN(salary) FROM employees e WHERE e.dept_id = d.id) as min_salary,
    (SELECT SUM(p.budget) FROM projects p WHERE p.dept_id = d.id AND p.status = 'active') as active_project_budget
FROM departments d
ORDER BY d.id;

-- Cleanup
DROP TABLE projects;
DROP TABLE employees;
DROP TABLE departments;