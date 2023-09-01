-- +goose Up
-- Table files contains all media files.
-- The table does not contain the actual file contents but only queryable metadata.
-- The actual file contents can be retrieved based on the UUID of the file.
CREATE TABLE files
(
    LIKE entity INCLUDING ALL,

    upload_id INTEGER REFERENCES uploads (id),
    path      TEXT     NOT NULL DEFAULT '',

    type      TEXT     NOT NULL DEFAULT '' CHECK ( type ~* '[^ \/]+\/[^ \/]+' ),
    size      INT8     NOT NULL DEFAULT 0,
    checksum  BYTEA    NOT NULL DEFAULT ''::BYTEA,

    -- Audio & Video
    duration  INTERVAL NOT NULL DEFAULT '0'::INTERVAL,

-- Videos & Images
    width     INT      NOT NULL DEFAULT 0,
    height    INT      NOT NULL DEFAULT 0
);

-- Trigger updated_at sets files.updated_at during updates.
CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON files
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON files;
DROP TABLE IF EXISTS files;
