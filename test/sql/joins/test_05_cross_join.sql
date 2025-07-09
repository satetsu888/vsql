-- Test 5: CROSS JOIN
-- Expected: 20 rows (4 users x 5 colors)
-- Test: Cartesian product of users and colors

-- Create test tables
CREATE TABLE users (
    id INTEGER,
    name TEXT
);

CREATE TABLE colors (
    id INTEGER,
    color TEXT
);

-- Insert test data
INSERT INTO users (id, name) VALUES
    (1, 'Alice'),
    (2, 'Bob'),
    (3, 'Charlie'),
    (4, 'Dave');

INSERT INTO colors (id, color) VALUES
    (1, 'Red'),
    (2, 'Blue'),
    (3, 'Green'),
    (4, 'Yellow'),
    (5, 'Purple');

-- Test query
SELECT u.name, c.color
FROM users u
CROSS JOIN colors c
ORDER BY u.name, c.color;

-- Cleanup
DROP TABLE colors;
DROP TABLE users;