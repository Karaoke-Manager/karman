-- +goose Up
-- Table uploads stores all uploads in the system.
-- An upload is associated with a storage location based on its uuid.
CREATE TABLE uploads
(
    LIKE entity INCLUDING ALL,

    open            bool NOT NULL DEFAULT TRUE,
    songs_total     INT  NOT NULL DEFAULT -1,
    songs_processed INT  NOT NULL DEFAULT -1
);

-- Table upload_errors stores errors that occurred during processing of an upload.
-- An upload_error is always linked to an upload, so it is not created from the entity table.
CREATE TABLE upload_errors
(
    id        INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    file      TEXT NOT NULL,
    message   TEXT NOT NULL,

    upload_id INT  NOT NULL REFERENCES uploads (id) ON DELETE CASCADE
);

-- Trigger updated_at sets uploads.updated_at during updates.
CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON uploads
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON uploads;
DROP TABLE IF EXISTS upload_errors;
DROP TABLE IF EXISTS uploads;
