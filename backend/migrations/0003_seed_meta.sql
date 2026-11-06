-- Метаданные развёртывания, нужные для паритета ответов mock API.
--
-- Ответ GET /api/page содержит каталог данных и имена файлов фрагментов,
-- а /api/health — число секций главной страницы. Это свойства развёртывания,
-- а не бизнес-данные, поэтому они хранятся отдельной таблицей «ключ → JSON».

BEGIN;

CREATE TABLE seed_meta (
    key        text PRIMARY KEY,
    value      text NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now()
);

COMMENT ON TABLE seed_meta IS 'Метаданные загрузки данных: каталоги фрагментов, счётчики.';

INSERT INTO schema_migration (version) VALUES ('0003_seed_meta');

COMMIT;
