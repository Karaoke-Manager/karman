-- +goose Up
CREATE TABLE uploads
(
    LIKE entity INCLUDING ALL,

    open            bool NOT NULL DEFAULT TRUE,
    songs_total     INT  NOT NULL DEFAULT -1,
    songs_processed INT  NOT NULL DEFAULT -1
);


CREATE TABLE upload_errors
(
    id        INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    file      TEXT NOT NULL,
    message   TEXT NOT NULL,

    upload_id INT  NOT NULL REFERENCES uploads (id)
);

CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON uploads
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON uploads;
DROP TABLE IF EXISTS upload_errors;
DROP TABLE IF EXISTS uploads;
