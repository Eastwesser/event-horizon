-- Написать select который сделает OUT

CREATE TABLE test(
                     id SERIAL,
                     name VARCHAR
)
если сделать обычный select, то получим:
                         | id | name  |
                         | 1  | name1 |
                         | 2  | name2 |
                         | 3  | name3 |
                         | 4  | name4 |
                         | 5  | name5 |
                         OUT - то как надо вывести
                         | id | name  |
                         | 1  | name1 |
                         | 3  | name3 |
                         | 2  | name2 |
                         | 4  | name4 |
                         | 5  | name5 |