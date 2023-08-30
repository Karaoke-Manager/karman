-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE entity
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);


-- +goose StatementBegin
CREATE FUNCTION tg_set_updated_at()
    RETURNS TRIGGER
AS
$$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS tg_set_updated_at();
DROP TABLE IF EXISTS entity;
