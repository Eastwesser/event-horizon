CREATE TABLE holds (
                       id         TEXT PRIMARY KEY,
                       account_id BIGINT NOT NULL FK accounts(id),
                       amount     BIGINT NOT NULL
);

CREATE TABLE accounts (
                          id      BIGSERIAL PRIMARY KEY,
                          amount BIGINT NOT NULL,  -- баланс без учета холдов
);
-- POST /accounts/:id/hold?amount=10
BEGIN;

SELECT amount FROM accounts WHERE id = :id;
.........
   UPDATE accounts SET amount = 90;
INSERT INTO holds;

COMMIT;ROLLBACK;
-- какие могут быть проблемы, как исправить?