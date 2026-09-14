-- Дали три таблицы - юзеры, покупки, банлист. по ним надо было написать два запроса.
-- Первый должен был выбрать юзеров и айди покупки, которую они совершили до того,
-- как их забанили, либо если они вообще не забанены.
-- Второй запрос - надо было выбрать юзеров, которые совершили покупки суммарно больше, чем на 5к

CREATE TABLE users (
                       user_id SERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL
);


CREATE TABLE purchases (
                           purchase_id SERIAL PRIMARY KEY,
                           user_id INT NOT NULL,
                           amount DECIMAL(10, 2) NOT NULL,
                           purchase_date TIMESTAMP NOT NULL,
                           FOREIGN KEY (user_id) REFERENCES users(user_id)
);


CREATE TABLE banlist (
                         user_id INT PRIMARY KEY,
                         ban_date TIMESTAMP NOT NULL,
                         FOREIGN KEY (user_id) REFERENCES users(user_id)
);