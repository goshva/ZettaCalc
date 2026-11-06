-- Роль пользователя: нужна, чтобы различать клиента и менеджера.
--
-- Роль участвует в проверке прав (code-based evals), поэтому она вынесена в
-- отдельную колонку со строгим списком значений, а не хранится в свободном
-- поле. Значение по умолчанию — client: демо-пользователь mock API клиентский.

BEGIN;

ALTER TABLE app_user
    ADD COLUMN role text NOT NULL DEFAULT 'client'
        CHECK (role IN ('client', 'manager', 'admin'));

COMMENT ON COLUMN app_user.role IS 'Роль пользователя: client | manager | admin.';

INSERT INTO schema_migration (version) VALUES ('0002_app_user_role');

COMMIT;
