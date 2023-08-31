-- +goose Up
-- +goose StatementBegin
-- Function notes_lyrics calculates the lyrics for a string of notes.
-- The input is expected to be in the UltraStar TXT format.
-- The output is a text containing the complete lyrics.
-- Multiplayer lyrics are not supported by this function.
CREATE FUNCTION notes_lyrics(s TEXT)
    RETURNS TEXT
    RETURNS NULL ON NULL INPUT
AS
$$
BEGIN
    -- Remove empty lines
    s := REGEXP_REPLACE(s, '^\s*$', '', 'gn');
    -- Remove lines that do not affect lyrics, such as BPM changes.
    s := REGEXP_REPLACE(s, '^[^:*FRG-](.*)(\n|\Z)', '', 'gn');
    -- Line Breaks
    s := REGEXP_REPLACE(s, '^-(.*)$', '', 'gn');
    -- Lyric Syllables
    s := REGEXP_REPLACE(s, '^[:*FRG]\s*(-?\d+)\s*(-?\d+)\s*(-?\d+)\s(.*)(\n|\Z)', '\4', 'gn');
    RETURN s;
END;
$$ LANGUAGE plpgsql
    IMMUTABLE
    PARALLEL SAFE;
-- +goose StatementEnd

-- Table songs stores all Karman songs.
CREATE TABLE songs
(
    LIKE entity INCLUDING ALL,

    upload_id          INTEGER REFERENCES uploads (id) ON DELETE CASCADE,
    audio_file_id      INTEGER  REFERENCES files (id) ON DELETE SET NULL,
    cover_file_id      INTEGER  REFERENCES files (id) ON DELETE SET NULL,
    video_file_id      INTEGER  REFERENCES files (id) ON DELETE SET NULL,
    background_file_id INTEGER  REFERENCES files (id) ON DELETE SET NULL,

    bpm                FLOAT    NOT NULL DEFAULT 0,
    gap                INTERVAL NOT NULL DEFAULT '0'::INTERVAL,
    video_gap          INTERVAL NOT NULL DEFAULT '0'::INTERVAL,
    start              INTERVAL NOT NULL DEFAULT '0'::INTERVAL,
    "end"              INTERVAL NOT NULL DEFAULT '0'::INTERVAL,
    preview_start      INTERVAL NOT NULL DEFAULT '0'::INTERVAL,
    medley_start_beat  INTEGER  NOT NULL DEFAULT 0,
    medley_end_beat    INTEGER  NOT NULL DEFAULT 0,
    manual_medley      BOOLEAN  NOT NULL DEFAULT FALSE,

    title              TEXT     NOT NULL DEFAULT '',
    artists            TEXT[]   NOT NULL DEFAULT '{}'::TEXT[],
    genre              TEXT     NOT NULL DEFAULT '',
    edition            TEXT     NOT NULL DEFAULT '',
    creator            TEXT     NOT NULL DEFAULT '',
    language           TEXT     NOT NULL DEFAULT '',
    year               INT      NOT NULL DEFAULT 0,
    comment            TEXT     NOT NULL DEFAULT '',
    extra              JSONB    NOT NULL DEFAULT '{}'::JSONB,

    notes_p1           TEXT     NOT NULL DEFAULT '',
    notes_p2           TEXT,
    duet_singer1       TEXT     NOT NULL DEFAULT '',
    duet_singer2       TEXT     NOT NULL DEFAULT '',

    lyrics_p1          TEXT GENERATED ALWAYS AS ( notes_lyrics(notes_p1) ) STORED,
    lyrics_p2          TEXT GENERATED ALWAYS AS ( notes_lyrics(notes_p2) ) STORED,
    is_duet            BOOLEAN GENERATED ALWAYS AS ( notes_p2 IS NOT NULL ) STORED
);

-- Trigger updated_at sets songs.updated_at during updates.
CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON songs
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();


-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON songs;
DROP TABLE IF EXISTS songs;
DROP FUNCTION IF EXISTS notes_lyrics;
