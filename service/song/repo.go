package song

import (
	"context"
	"errors"
	"time"

	"codello.dev/ultrastar"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgxutil"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/service/dbutil"
)

// dbRepo is the main Repository implementation, backed by a PostgreSQL database.
type dbRepo struct {
	db pgxutil.DB // database connection
}

// NewDBRepository creates a new Repository backed by the specified database connection.
// db can be a single connection or a connection pool.
func NewDBRepository(db pgxutil.DB) Repository {
	return &dbRepo{db}
}

// songRow is the data returned by a SELECT query for songs.
// This type is used by GetSong and FindSongs.
type songRow struct {
	UUID      uuid.UUID
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at"`

	BPM             ultrastar.BPM
	Gap             time.Duration
	VideoGap        time.Duration `db:"video_gap"`
	Start           time.Duration
	End             time.Duration
	PreviewStart    time.Duration  `db:"preview_start"`
	MedleyStartBeat ultrastar.Beat `db:"medley_start_beat"`
	MedleyEndBeat   ultrastar.Beat `db:"medley_end_beat"`
	ManualMedley    bool           `db:"manual_medley"`

	Title       string
	Artists     []string
	Genre       string
	Edition     string
	Creator     string
	Language    string
	Year        int
	Comment     string
	Extra       map[string]string
	DuetSinger1 string `db:"duet_singer1"`
	DuetSinger2 string `db:"duet_singer2"`

	NotesP1 dbutil.Notes `db:"notes_p1"`
	NotesP2 dbutil.Notes `db:"notes_p2"`

	AudioUUID      uuid.NullUUID        `db:"audio_uuid"`
	AudioCreatedAt pgtype.Timestamp     `db:"audio_created_at"`
	AudioUpdatedAt pgtype.Timestamp     `db:"audio_updated_at"`
	AudioDeletedAt pgtype.Timestamp     `db:"audio_deleted_at"`
	AudioType      *mediatype.MediaType `db:"audio_type"`
	AudioSize      pgtype.Int8          `db:"audio_size"`
	AudioChecksum  []byte               `db:"audio_checksum"`
	AudioDuration  *time.Duration       `db:"audio_duration"`

	CoverUUID      uuid.NullUUID        `db:"cover_uuid"`
	CoverCreatedAt pgtype.Timestamp     `db:"cover_created_at"`
	CoverUpdatedAt pgtype.Timestamp     `db:"cover_updated_at"`
	CoverDeletedAt pgtype.Timestamp     `db:"cover_deleted_at"`
	CoverType      *mediatype.MediaType `db:"cover_type"`
	CoverSize      pgtype.Int8          `db:"cover_size"`
	CoverChecksum  []byte               `db:"cover_checksum"`
	CoverWidth     pgtype.Int4          `db:"cover_width"`
	CoverHeight    pgtype.Int4          `db:"cover_height"`

	VideoUUID      uuid.NullUUID        `db:"video_uuid"`
	VideoCreatedAt pgtype.Timestamp     `db:"video_created_at"`
	VideoUpdatedAt pgtype.Timestamp     `db:"video_updated_at"`
	VideoDeletedAt pgtype.Timestamp     `db:"video_deleted_at"`
	VideoType      *mediatype.MediaType `db:"video_type"`
	VideoSize      pgtype.Int8          `db:"video_size"`
	VideoChecksum  []byte               `db:"video_checksum"`
	VideoDuration  *time.Duration       `db:"video_duration"`
	VideoWidth     pgtype.Int4          `db:"video_width"`
	VideoHeight    pgtype.Int4          `db:"video_height"`

	BackgroundUUID      uuid.NullUUID        `db:"bg_uuid"`
	BackgroundCreatedAt pgtype.Timestamp     `db:"bg_created_at"`
	BackgroundUpdatedAt pgtype.Timestamp     `db:"bg_updated_at"`
	BackgroundDeletedAt pgtype.Timestamp     `db:"bg_deleted_at"`
	BackgroundType      *mediatype.MediaType `db:"bg_type"`
	BackgroundSize      pgtype.Int8          `db:"bg_size"`
	BackgroundChecksum  []byte               `db:"bg_checksum"`
	BackgroundWidth     pgtype.Int4          `db:"bg_width"`
	BackgroundHeight    pgtype.Int4          `db:"bg_height"`
}

// toModel converts r into an equivalent model.Song.
func (r songRow) toModel() model.Song {
	song := model.Song{
		Model: model.Model{
			UUID:      r.UUID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		Artists: r.Artists,
		Song: ultrastar.Song{
			BPM:             r.BPM,
			Gap:             r.Gap,
			VideoGap:        r.VideoGap,
			Start:           r.Start,
			End:             r.End,
			PreviewStart:    r.PreviewStart,
			MedleyStartBeat: r.MedleyStartBeat,
			MedleyEndBeat:   r.MedleyEndBeat,
			NoAutoMedley:    r.ManualMedley,
			Title:           r.Title,
			Genre:           r.Genre,
			Edition:         r.Edition,
			Creator:         r.Creator,
			Language:        r.Language,
			Year:            r.Year,
			Comment:         r.Comment,
			CustomTags:      r.Extra,
			DuetSinger1:     r.DuetSinger1,
			DuetSinger2:     r.DuetSinger2,
			NotesP1:         ultrastar.Notes(r.NotesP1),
			NotesP2:         ultrastar.Notes(r.NotesP2),
		}}
	if r.DeletedAt.Valid {
		song.DeletedAt = r.DeletedAt.Time
	}
	if r.AudioUUID.Valid {
		song.AudioFile = &model.File{
			Model: model.Model{
				UUID:      r.AudioUUID.UUID,
				CreatedAt: r.AudioCreatedAt.Time,
				UpdatedAt: r.AudioUpdatedAt.Time,
			},
			Type:     dbutil.ZeroNil(r.AudioType),
			Size:     r.AudioSize.Int64,
			Checksum: r.AudioChecksum,
			Duration: dbutil.ZeroNil(r.AudioDuration),
		}
		if r.AudioDeletedAt.Valid {
			song.AudioFile.DeletedAt = r.AudioDeletedAt.Time
		}
	}
	if r.CoverUUID.Valid {
		song.CoverFile = &model.File{
			Model: model.Model{
				UUID:      r.CoverUUID.UUID,
				CreatedAt: r.CoverCreatedAt.Time,
				UpdatedAt: r.CoverUpdatedAt.Time,
			},
			Type:     dbutil.ZeroNil(r.CoverType),
			Size:     r.CoverSize.Int64,
			Checksum: r.CoverChecksum,
			Width:    int(r.CoverWidth.Int32),
			Height:   int(r.CoverHeight.Int32),
		}
		if r.CoverDeletedAt.Valid {
			song.CoverFile.DeletedAt = r.CoverDeletedAt.Time
		}
	}
	if r.VideoUUID.Valid {
		song.VideoFile = &model.File{
			Model: model.Model{
				UUID:      r.VideoUUID.UUID,
				CreatedAt: r.VideoCreatedAt.Time,
				UpdatedAt: r.VideoUpdatedAt.Time,
			},
			Type:     dbutil.ZeroNil(r.VideoType),
			Size:     r.VideoSize.Int64,
			Checksum: r.VideoChecksum,
			Duration: dbutil.ZeroNil(r.VideoDuration),
			Width:    int(r.VideoWidth.Int32),
			Height:   int(r.VideoHeight.Int32),
		}
		if r.VideoDeletedAt.Valid {
			song.VideoFile.DeletedAt = r.VideoDeletedAt.Time
		}
	}
	if r.BackgroundUUID.Valid {
		song.BackgroundFile = &model.File{
			Model: model.Model{
				UUID:      r.BackgroundUUID.UUID,
				CreatedAt: r.BackgroundCreatedAt.Time,
				UpdatedAt: r.BackgroundUpdatedAt.Time,
			},
			Type:     dbutil.ZeroNil(r.BackgroundType),
			Size:     r.BackgroundSize.Int64,
			Checksum: r.BackgroundChecksum,
			Width:    int(r.BackgroundWidth.Int32),
			Height:   int(r.BackgroundHeight.Int32),
		}
		if r.BackgroundDeletedAt.Valid {
			song.BackgroundFile.DeletedAt = r.BackgroundDeletedAt.Time
		}
	}
	return song
}

// CreateSong creates song in the database.
// This method also sets nil values in song to equivalent zero values in order to avoid constraint violations.
func (r *dbRepo) CreateSong(ctx context.Context, song *model.Song) error {
	prepareSong(song)
	row, err := pgxutil.InsertRowReturning(ctx, r.db, "songs", map[string]any{
		"title":    song.Title,
		"artists":  song.Artists,
		"genre":    song.Genre,
		"edition":  song.Edition,
		"creator":  song.Creator,
		"language": song.Language,
		"year":     song.Year,
		"comment":  song.Comment,
		"extra":    song.CustomTags,

		"bpm":               song.BPM,
		"gap":               song.Gap,
		"video_gap":         song.VideoGap,
		"start":             song.Start,
		"end":               song.End,
		"preview_start":     song.PreviewStart,
		"medley_start_beat": song.MedleyStartBeat,
		"medley_end_beat":   song.MedleyEndBeat,
		"manual_medley":     song.NoAutoMedley,

		"notes_p1":     dbutil.Notes(song.NotesP1),
		"notes_p2":     dbutil.Notes(song.NotesP2),
		"duet_singer1": song.DuetSinger1,
		"duet_singer2": song.DuetSinger2,
	}, "uuid, created_at, updated_at", pgx.RowToStructByName[struct {
		UUID      uuid.UUID
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}])
	if err != nil {
		return err
	}
	song.UUID = row.UUID
	song.CreatedAt = row.CreatedAt
	song.UpdatedAt = row.UpdatedAt
	return nil
}

// prepareSong modifies song in a way that it can be inserted into the database.
// This mainly concerns replacing nil values with non-nil zero values.
func prepareSong(song *model.Song) {
	if song.Artists == nil {
		song.Artists = make([]string, 0)
	}
	if song.CustomTags == nil {
		song.CustomTags = make(map[string]string)
	}
	if song.NotesP1 == nil {
		song.NotesP1 = make(ultrastar.Notes, 0)
	}
}

// GetSong fetches a single song from the database by its UUID.
func (r *dbRepo) GetSong(ctx context.Context, id uuid.UUID) (model.Song, error) {
	row, err := pgxutil.SelectRow(ctx, r.db, `SELECT
    s.uuid, s.created_at, s.updated_at, s.deleted_at,
    s.title, s.artists, s.genre, s.edition, s.creator, s.language, s.year, s.comment, s.extra,
    s.bpm, s.gap, s.video_gap, s.start, s."end", s.preview_start, s.medley_start_beat, s.medley_end_beat, s.manual_medley,
    s.duet_singer1, s.duet_singer2, s.notes_p1, s.notes_p2,
    
    a.uuid AS audio_uuid, a.created_at AS audio_created_at, a.updated_at AS audio_updated_at, a.deleted_at AS audio_deleted_at, a.type AS audio_type, a.size AS audio_size, a.checksum AS audio_checksum, a.duration AS audio_duration,
    c.uuid AS cover_uuid, c.created_at AS cover_created_at, c.updated_at AS cover_updated_at, c.deleted_at AS cover_deleted_at, c.type AS cover_type, c.size AS cover_size, c.checksum AS cover_checksum, c.width cover_width, c.height AS cover_height,
    v.uuid AS video_uuid, v.created_at AS video_created_at, v.updated_at AS video_updated_at, v.deleted_at AS video_deleted_at, v.type AS video_type, v.size AS video_size, v.checksum AS video_checksum, v.duration AS video_duration, v.width AS video_width, v.height AS video_height,
    b.uuid AS bg_uuid, b.created_at AS bg_created_at, b.updated_at AS bg_updated_at, b.deleted_at AS bg_deleted_at, b.type AS bg_type, b.size AS bg_size, b.checksum AS bg_checksum, b.width AS bg_width, b.height AS bg_height
    FROM songs AS s
        LEFT OUTER JOIN files AS a ON s.audio_file_id = a.id
    	LEFT OUTER JOIN files AS c ON s.cover_file_id = c.id
    	LEFT OUTER JOIN files AS v ON s.video_file_id = v.id
    	LEFT OUTER JOIN files AS b ON s.background_file_id = b.id
    WHERE S.uuid = $1`, []any{id}, pgx.RowToStructByName[songRow])
	if err != nil {
		return model.Song{}, dbutil.Error(err)
	}
	return row.toModel(), nil
}

// FindSongs fetches multiple songs from the database.
// The results are paginated with limit and offset.
func (r *dbRepo) FindSongs(ctx context.Context, limit int, offset int64) ([]model.Song, int64, error) {
	total, err := pgxutil.SelectRow(ctx, r.db, `SELECT COUNT(*)
	FROM songs AS s
	WHERE s.upload_id IS NULL`, nil, pgx.RowTo[int64])
	if err != nil {
		return nil, 0, err
	}

	songs, err := pgxutil.Select(ctx, r.db, `SELECT
    s.uuid, s.created_at, s.updated_at, s.deleted_at,
    s.title, s.artists, s.genre, s.edition, s.creator, s.language, s.year, s.comment, s.extra,
    s.bpm, s.gap, s.video_gap, s.start, s."end", s.preview_start, s.medley_start_beat, s.medley_end_beat, s.manual_medley,
    s.notes_p1, s.notes_p2, s.duet_singer1, s.duet_singer2,
    
    a.uuid AS audio_uuid, a.created_at AS audio_created_at, a.updated_at AS audio_updated_at, a.deleted_at AS audio_deleted_at, a.type AS audio_type, a.size AS audio_size, a.checksum AS audio_checksum, a.duration AS audio_duration,
    c.uuid AS cover_uuid, c.created_at AS cover_created_at, c.updated_at AS cover_updated_at, c.deleted_at AS cover_deleted_at, c.type AS cover_type, c.size AS cover_size, c.checksum AS cover_checksum, c.width cover_width, c.height AS cover_height,
    v.uuid AS video_uuid, v.created_at AS video_created_at, v.updated_at AS video_updated_at, v.deleted_at AS video_deleted_at, v.type AS video_type, v.size AS video_size, v.checksum AS video_checksum, v.duration AS video_duration, v.width AS video_width, v.height AS video_height,
    b.uuid AS bg_uuid, b.created_at AS bg_created_at, b.updated_at AS bg_updated_at, b.deleted_at AS bg_deleted_at, b.type AS bg_type, b.size AS bg_size, b.checksum AS bg_checksum, b.width AS bg_width, b.height AS bg_height
    FROM songs AS s
        LEFT OUTER JOIN files AS a ON s.audio_file_id = a.id
    	LEFT OUTER JOIN files AS c ON s.cover_file_id = c.id
    	LEFT OUTER JOIN files AS v ON s.video_file_id = v.id
    	LEFT OUTER JOIN files AS b ON s.background_file_id = b.id
	WHERE s.upload_id IS NULL
	LIMIT CASE WHEN $1 < 0 THEN NULL ELSE $1 END OFFSET $2`, []any{limit, offset}, func(row pgx.CollectableRow) (model.Song, error) {
		data, err := pgx.RowToStructByName[songRow](row)
		if err != nil {
			return model.Song{}, err
		}
		return data.toModel(), nil
	})
	return songs, total, err
}

// UpdateSong updates the song in the database with song.UUID.
// File references must already exist in the database or they will be set to nil.
// Data of file references (size, checksum, ...) is not updated.
func (r *dbRepo) UpdateSong(ctx context.Context, song *model.Song) error {
	prepareSong(song)
	var audioUUID, coverUUID, videoUUID, backgroundUUID uuid.NullUUID
	if song.AudioFile != nil {
		audioUUID = uuid.NullUUID{UUID: song.AudioFile.UUID, Valid: true}
	}
	if song.CoverFile != nil {
		coverUUID = uuid.NullUUID{UUID: song.CoverFile.UUID, Valid: true}
	}
	if song.VideoFile != nil {
		videoUUID = uuid.NullUUID{UUID: song.VideoFile.UUID, Valid: true}
	}
	if song.BackgroundFile != nil {
		backgroundUUID = uuid.NullUUID{UUID: song.BackgroundFile.UUID, Valid: true}
	}
	row, err := pgxutil.SelectRow(ctx, r.db, `UPDATE songs SET 
		title = $2, artists = $3, genre = $4, edition = $5, creator = $6, language = $7, year = $8, comment = $9, extra = $10,
		bpm = $11, gap = $12, video_gap = $13, start = $14, "end" = $15, preview_start = $16, medley_start_beat = $17, medley_end_beat = $18, manual_medley = $19,
		notes_p1 = $20, notes_p2 = $21, duet_singer1 = $22, duet_singer2 = $23,
		audio_file_id = CASE WHEN $24::uuid IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $24) END,
		cover_file_id = CASE WHEN $25::uuid IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $25) END,
		video_file_id = CASE WHEN $26::uuid IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $26) END,
		background_file_id = CASE WHEN $27::uuid IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $27) END
	WHERE uuid = $1
	RETURNING updated_at, audio_file_id, cover_file_id, video_file_id, background_file_id`, []any{
		song.UUID,
		song.Title, song.Artists, song.Genre, song.Edition, song.Creator, song.Language, song.Year, song.Comment, song.CustomTags,
		song.BPM, song.Gap, song.VideoGap, song.Start, song.End, song.PreviewStart, song.MedleyStartBeat, song.MedleyEndBeat, song.NoAutoMedley,
		dbutil.Notes(song.NotesP1), dbutil.Notes(song.NotesP2), song.DuetSinger1, song.DuetSinger2,
		audioUUID, coverUUID, videoUUID, backgroundUUID,
	}, pgx.RowToStructByName[struct {
		UpdatedAt        time.Time   `db:"updated_at"`
		AudioFileID      pgtype.Int4 `db:"audio_file_id"`
		CoverFileID      pgtype.Int4 `db:"cover_file_id"`
		VideoFileID      pgtype.Int4 `db:"video_file_id"`
		BackgroundFileID pgtype.Int4 `db:"background_file_id"`
	}])
	if err != nil {
		return dbutil.Error(err)
	}
	song.UpdatedAt = row.UpdatedAt
	if !row.AudioFileID.Valid {
		song.AudioFile = nil
	}
	if !row.CoverFileID.Valid {
		song.CoverFile = nil
	}
	if !row.VideoFileID.Valid {
		song.VideoFile = nil
	}
	if !row.BackgroundFileID.Valid {
		song.BackgroundFile = nil
	}
	return nil
}

// DeleteSong deletes the song with the specified UUID from the database.
// If no song with the specified UUID existed, the first return value will be false.
func (r *dbRepo) DeleteSong(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, r.db, `DELETE FROM songs WHERE uuid = $1`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
