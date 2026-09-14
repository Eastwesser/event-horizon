CREATE TABLE a_user (
    id
    user_id
    
}

SELECT a.id, b.queries
FROM a_user a
JOIN b_queries b
    ON a.id = b.user_id
WHERE b.queries > 100;