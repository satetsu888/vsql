-- Test 30: Count after bulk insert
-- Expected: 1 row, count=17

-- Setup
CREATE TABLE users (id int, name text, email text);
INSERT INTO users (id, name, email) VALUES (1, 'Alice', 'alice@example.com');
INSERT INTO users (id, name, age) VALUES (2, 'Bob', 30);
INSERT INTO users (id, name, email, country) VALUES (3, 'Charlie', 'charlie@example.com', 'USA');
INSERT INTO users (id, name, email) VALUES (4, NULL, NULL);
INSERT INTO users (id, name) VALUES (5, 'O''Brien');
INSERT INTO users (id, name) VALUES (6, 'Test!@#$%^&*()');
INSERT INTO users (id, name) VALUES (7, '日本語テスト');
INSERT INTO users (id, name, email) VALUES 
    (101, 'User101', 'user101@example.com'),
    (102, 'User102', 'user102@example.com'),
    (103, 'User103', 'user103@example.com'),
    (104, 'User104', 'user104@example.com'),
    (105, 'User105', 'user105@example.com'),
    (106, 'User106', 'user106@example.com'),
    (107, 'User107', 'user107@example.com'),
    (108, 'User108', 'user108@example.com'),
    (109, 'User109', 'user109@example.com'),
    (110, 'User110', 'user110@example.com');

-- Test Query
SELECT COUNT(*) FROM users;

-- Cleanup
DROP TABLE users;