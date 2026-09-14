-- Задание 1
-- Вывести все айтемы с максимальной стоимостью
create table example1
(
    id    serial,
    title text,
    cost  int
);
insert into example1(title, cost)
values ('item1', 10),
       ('item2', 12),
       ('item3', 15),
       ('item4', 10),
       ('item5', 15);
-- ===========================================
-- Задание 2
-- Вывести последние комментарии каждого автора(по дате)
create table example2
(
    id      serial,
    comment text,
    author  text,
    date    timestamp with time zone
);
insert into example2(comment, author, date)
values ('aaa', 'ivanov', '2024-03-18 09:45:53 +00:00'),
       ('bbb', 'petrov', '2024-03-16 07:24:53 +00:00'),
       ('ccc', 'ivanov', '2024-03-20 17:01:00 +00:00'),
       ('ddd', 'petrov', '2024-03-18 09:45:53 +00:00'),
       ('eee', 'sidorov', '2024-03-19 11:21:53 +00:00');