SELECT *
FROM test
ORDER BY CASE
             WHEN id = 1 THEN 1
             WHEN id = 3 THEN 2
             WHEN id = 2 THEN 3
             WHEN id = 4 THEN 4
             WHEN id = 5 THEN 5
             END;