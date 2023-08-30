-- +goose Up
CREATE TABLE files
(
    LIKE entity INCLUDING ALL,

    upload_id INTEGER REFERENCES uploads (id),
    path      TEXT     NOT NULL DEFAULT '',

    type      TEXT     NOT NULL DEFAULT '' CHECK ( type ~* '[^ \/]+\/[^ \/]+' ),
    size      INT8     NOT NULL DEFAULT 0,
    checksum  BYTEA    NOT NULL DEFAULT ''::bytea,

    -- Audio & Video
    duration  INTERVAL NOT NULL DEFAULT 0,

-- Videos & Images
    width     INT      NOT NULL DEFAULT 0,
    height    INT      NOT NULL DEFAULT 0
);


CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON files
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON files;
DROP TABLE IF EXISTS files;
