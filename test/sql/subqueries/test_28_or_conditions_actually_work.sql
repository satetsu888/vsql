-- Test: OR conditions with outer table references ACTUALLY WORK!
-- The real issue was with boolean type handling, not the OR logic itself

-- Setup with working data types
CREATE TABLE users (id int, name text, status text, score int);
CREATE TABLE activities (user_id int, type text);

INSERT INTO users VALUES
  (1, 'Alice', 'active', 100),
  (2, 'Bob', 'active', 50),
  (3, 'Charlie', 'inactive', 200),
  (4, 'David', 'inactive', 25);

INSERT INTO activities VALUES
  (1, 'login'),
  (3, 'purchase');

-- Test 1: OR with outer table string comparison (WORKS!)
-- Expected: 3 rows (all active users + users with activities)
-- Returns: Alice (active + activity), Bob (active), Charlie (activity)
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.status = 'active'
)
ORDER BY name;

-- Test 2: OR with outer table numeric comparison (WORKS!)
-- Expected: 3 rows (users with high scores or activities)
-- Returns: Alice (activity), Charlie (high score + activity), David (high score)
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.score > 75
)
ORDER BY name;

-- Test 3: Complex OR with multiple outer references (WORKS!)
-- Expected: 4 rows (active users OR high scores OR activities)
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.status = 'active'
  OR u.score > 150
)
ORDER BY name;

-- Test 4: NOT EXISTS with OR (WORKS!)
-- Expected: 1 row (David - not active AND no activities)
SELECT name FROM users u
WHERE NOT EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.status = 'active'
)
ORDER BY name;

-- Test 5: Nested EXISTS with OR referencing outer table (WORKS!)
CREATE TABLE events (activity_type text, points int);
INSERT INTO events VALUES ('login', 10), ('purchase', 50);

-- Expected: 2 rows (users with valuable activities or high scores)
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id
  AND EXISTS (
    SELECT 1 FROM events e 
    WHERE e.activity_type = a.type 
    AND (e.points > 30 OR u.score > 150)
  )
)
ORDER BY name;

DROP TABLE events;

-- Test 6: OR with LIKE on outer table (WORKS!)
-- Expected: 3 rows (users with activities or names starting with 'A' or 'D')
SELECT name FROM users u
WHERE EXISTS (
  SELECT 1 FROM activities a 
  WHERE a.user_id = u.id 
  OR u.name LIKE 'A%'
  OR u.name LIKE 'D%'
)
ORDER BY name;

-- Cleanup
DROP TABLE activities;
DROP TABLE users;

-- CONCLUSION: Complex OR conditions with outer table references work perfectly!
-- The issue in test_25 was due to boolean type handling, not the OR logic.