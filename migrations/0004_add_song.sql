-- +goose Up
-- +goose StatementBegin
CREATE FUNCTION song_lyrics(s TEXT)
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

CREATE TABLE songs
(
    LIKE entity INCLUDING ALL,

    upload_id          INTEGER REFERENCES uploads (id),
    audio_file_id      INTEGER REFERENCES files (id),
    cover_file_id      INTEGER REFERENCES files (id),
    video_file_id      INTEGER REFERENCES files (id),
    background_file_id INTEGER REFERENCES files (id),

    gap                INTERVAL NOT NULL DEFAULT 0,
    video_gap          INTERVAL NOT NULL DEFAULT 0,
    start              INTERVAL NOT NULL DEFAULT 0,
    "end"              INTERVAL NOT NULL DEFAULT 0,
    preview_start      INTERVAL NOT NULL DEFAULT 0,
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
    extra              JSONB    NOT NULL DEFAULT '{}'::jsonb,

    music_p1           TEXT     NOT NULL DEFAULT '',
    music_p2           TEXT,
    duet_singer1       TEXT     NOT NULL DEFAULT '',
    duet_singer2       TEXT     NOT NULL DEFAULT '',

    lyrics_p1          TEXT GENERATED ALWAYS AS ( song_lyrics(music_p1) ) STORED,
    lyrics_p2          TEXT GENERATED ALWAYS AS ( song_lyrics(music_p2) ) STORED,
    is_duet            BOOLEAN GENERATED ALWAYS AS ( music_p2 IS NOT NULL ) STORED
);

CREATE TRIGGER updated_at
    BEFORE UPDATE
    ON songs
    FOR EACH ROW
EXECUTE PROCEDURE tg_set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS updated_at ON songs;
DROP TABLE IF EXISTS songs;
DROP FUNCTION IF EXISTS song_lyrics;
