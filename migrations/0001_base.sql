-- +goose Up

-- Extension uuid-ossp is needed for UUID support.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Table entity is the base table for all other tables.
-- No data should be inserted into this table, it should only be used as a base
-- for other tables.
CREATE TABLE entity
(
    id         INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    uuid       UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP   NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);


-- +goose StatementBegin
-- Function tg_set_updated_at function updates the updated_at column to the current time.
-- This function is intended to be used as a trigger for tables created from the entity table.
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
