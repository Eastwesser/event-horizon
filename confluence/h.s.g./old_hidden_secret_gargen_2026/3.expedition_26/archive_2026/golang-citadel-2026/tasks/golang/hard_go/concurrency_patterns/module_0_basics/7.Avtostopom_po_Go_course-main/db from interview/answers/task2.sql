1) Создадим основные таблицы:
       - Автор (Authors)
       - Читатель (Readers)
       - Книга (Books)
       - Связь "Книга - Автор" (Book_Author),
         чтобы учесть отношение многие-ко-многим между Книгой и Автором.

    Важный момент:
    "Физически книга только одна и может быть только у одного читателя" —
    значит, в таблице Книга (Books) будет поле reader_id,
    указывающее на текущего читателя книги (NULL, если книга "на полке").

Примерная структура ниже.

Таблица с авторами

CREATE TABLE Authors (
    author_id   INT PRIMARY KEY AUTO_INCREMENT,
    author_name VARCHAR(100)   NOT NULL
);

Таблица с читателями

CREATE TABLE Readers (
    reader_id   INT PRIMARY KEY AUTO_INCREMENT,
    reader_name VARCHAR(100)   NOT NULL
);

Таблица с книгами

CREATE TABLE Books (
    book_id   INT PRIMARY KEY AUTO_INCREMENT,
    title     VARCHAR(255) NOT NULL,
    -- Поле, ссылающееся на читателя, у которого книга сейчас находится:
    reader_id INT NULL,
    FOREIGN KEY (reader_id) REFERENCES Readers(reader_id)
);

Связующая таблица для отношения многие-ко-многим "Книга - Автор"

CREATE TABLE Book_Author (
    id        INT PRIMARY KEY AUTO_INCREMENT,
    book_id   INT NOT NULL,
    author_id INT NOT NULL,
    FOREIGN KEY (book_id)   REFERENCES Books(book_id),
    FOREIGN KEY (author_id) REFERENCES Authors(author_id)
);

2) Запрос 1: выбрать названия всех книг, которые сейчас на руках
	(т.е. у которых reader_id не NULL)

SELECT b.title
FROM Books b
WHERE b.reader_id IS NOT NULL;

3) Запрос 2: выбрать названия всех книг, которые в библиотеке (reader_id IS NULL)
	и у которых больше 3 авторов

SELECT b.title
FROM Books b
JOIN Book_Author ba ON b.book_id = ba.book_id
WHERE b.reader_id IS NULL
GROUP BY b.book_id, b.title
HAVING COUNT(ba.author_id) > 3;

4) Запрос 3: выбрать имена топ-3 читаемых авторов (на данный момент),
       подразумеваем, что "читаемые" — значит книги в руках у читателей.
       Считаем количество взятых книг по каждому автору и сортируем по убыванию

SELECT a.author_name,
       COUNT(*) AS borrowed_count
FROM Authors a
JOIN Book_Author ba ON a.author_id = ba.author_id
JOIN Books b       ON b.book_id   = ba.book_id
WHERE b.reader_id IS NOT NULL
GROUP BY a.author_id, a.author_name
ORDER BY borrowed_count DESC
LIMIT 3;