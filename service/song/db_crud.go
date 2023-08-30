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

func coalesceZero[T any](p *T) T {
	if p != nil {
		return *p
	}
	var zero T
	return zero
}

type songRow struct {
	UUID      uuid.UUID
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
	DeletedAt pgtype.Timestamp `db:"deleted_at"`

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

	MusicP1 *dbutil.Music `db:"music_p1"`
	MusicP2 *dbutil.Music `db:"music_p2"`

	AudioUUID      uuid.NullUUID        `db:"a.uuid"`
	AudioCreatedAt pgtype.Timestamp     `db:"a.created_at"`
	AudioUpdatedAt pgtype.Timestamp     `db:"a.updated_at"`
	AudioDeletedAt pgtype.Timestamp     `db:"a.deleted_at"`
	AudioType      *mediatype.MediaType `db:"a.type"`
	AudioSize      pgtype.Int8          `db:"a.size"`
	AudioChecksum  []byte               `db:"a.checksum"`
	AudioDuration  *time.Duration       `db:"a.duration"`

	CoverUUID      uuid.NullUUID        `db:"c.uuid"`
	CoverCreatedAt pgtype.Timestamp     `db:"c.created_at"`
	CoverUpdatedAt pgtype.Timestamp     `db:"c.updated_at"`
	CoverDeletedAt pgtype.Timestamp     `db:"c.deleted_at"`
	CoverType      *mediatype.MediaType `db:"c.type"`
	CoverSize      pgtype.Int8          `db:"c.size"`
	CoverChecksum  []byte               `db:"c.checksum"`
	CoverWidth     pgtype.Int4          `db:"c.width"`
	CoverHeight    pgtype.Int4          `db:"c.height"`

	VideoUUID      uuid.NullUUID        `db:"v.uuid"`
	VideoCreatedAt pgtype.Timestamp     `db:"v.created_at"`
	VideoUpdatedAt pgtype.Timestamp     `db:"v.updated_at"`
	VideoDeletedAt pgtype.Timestamp     `db:"v.deleted_at"`
	VideoType      *mediatype.MediaType `db:"v.type"`
	VideoSize      pgtype.Int8          `db:"v.size"`
	VideoChecksum  []byte               `db:"v.checksum"`
	VideoDuration  *time.Duration       `db:"v.duration"`
	VideoWidth     pgtype.Int4          `db:"v.width"`
	VideoHeight    pgtype.Int4          `db:"v.height"`

	BackgroundUUID      uuid.NullUUID        `db:"b.uuid"`
	BackgroundCreatedAt pgtype.Timestamp     `db:"b.created_at"`
	BackgroundUpdatedAt pgtype.Timestamp     `db:"b.updated_at"`
	BackgroundDeletedAt pgtype.Timestamp     `db:"b.deleted_at"`
	BackgroundType      *mediatype.MediaType `db:"b.type"`
	BackgroundSize      pgtype.Int8          `db:"b.size"`
	BackgroundChecksum  []byte               `db:"b.checksum"`
	BackgroundWidth     pgtype.Int4          `db:"b.width"`
	BackgroundHeight    pgtype.Int4          `db:"b.height"`
}

func (r songRow) ToModel() *model.Song {
	song := &model.Song{
		Model: model.Model{
			UUID:      r.UUID,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		},
		Artists: r.Artists,
		Song: &ultrastar.Song{
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
			MusicP1:         (*ultrastar.Music)(r.MusicP1),
			MusicP2:         (*ultrastar.Music)(r.MusicP2),
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
			Type:     coalesceZero(r.AudioType),
			Size:     r.AudioSize.Int64,
			Checksum: r.AudioChecksum,
			Duration: coalesceZero(r.AudioDuration),
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
			Type:     coalesceZero(r.CoverType),
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
			Type:     coalesceZero(r.VideoType),
			Size:     r.VideoSize.Int64,
			Checksum: r.VideoChecksum,
			Duration: coalesceZero(r.VideoDuration),
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
			Type:     coalesceZero(r.BackgroundType),
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

func (db *db) CreateSong(ctx context.Context, song *model.Song) error {
	row, err := pgxutil.InsertRowReturning(ctx, db.q, "songs", map[string]any{
		"title":    song.Title,
		"artists":  song.Artists,
		"genre":    song.Genre,
		"edition":  song.Edition,
		"creator":  song.Creator,
		"language": song.Language,
		"year":     song.Year,
		"comment":  song.Comment,
		"extra":    song.CustomTags,

		"gap":               song.Gap,
		"video_gap":         song.VideoGap,
		"start":             song.Start,
		"end":               song.End,
		"preview_start":     song.PreviewStart,
		"medley_start_beat": song.MedleyStartBeat,
		"medley_end_beat":   song.MedleyEndBeat,
		"manual_medley":     song.NoAutoMedley,

		"music_p1":     (*dbutil.Music)(song.MusicP1),
		"music_p2":     (*dbutil.Music)(song.MusicP2),
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

func (db *db) GetSong(ctx context.Context, id uuid.UUID) (*model.Song, error) {
	row, err := pgxutil.SelectRow(ctx, db.q, `SELECT
    s.uuid, s.created_at, s.updated_at, s.deleted_at,
    s.gap, s.video_gap, s.start, s."end", s.preview_start, s.medley_start_beat, s.medley_end_beat, s.manual_medley,
    s.title, s.artists, s.genre, s.edition, s.creator, s.language, s.year, s.comment, s.extra,
    s.duet_singer1, s.duet_singer2, s.music_p1, s.music_p2,
    
    a.uuid, a.created_at, a.updated_at, a.deleted_at, a.type, a.size, a.checksum, a.duration,
    c.uuid, c.created_at, c.updated_at, c.deleted_at, c.type, c.size, c.checksum, c.width, c.height,
    v.uuid, v.created_at, v.updated_at, v.deleted_at, v.type, v.size, v.checksum, v.duration, v.width, v.height,
    b.uuid, b.created_at, b.updated_at, b.deleted_at, b.type, b.size, b.checksum, b.width, b.height
    FROM songs AS s
        LEFT OUTER JOIN files AS a ON s.audio_file_id = a.id
    	LEFT OUTER JOIN files AS c ON s.cover_file_id = c.id
    	LEFT OUTER JOIN files AS v ON s.video_file_id = v.id
    	LEFT OUTER JOIN files AS b ON s.background_file_id = b.id
    WHERE S.uuid = $1`, []any{id}, pgx.RowToStructByName[songRow])
	if err != nil {
		return nil, dbutil.Error(err)
	}
	return row.ToModel(), nil
}

func (db *db) FindSongs(ctx context.Context, limit int, offset int64) ([]*model.Song, int64, error) {
	total, err := pgxutil.SelectRow(ctx, db.q, `SELECT
    COUNT(*) OVER ()
	FROM songs AS s
	WHERE s.upload_id IS NULL`, nil, pgx.RowTo[int64])
	if err != nil {
		return nil, 0, err
	}

	songs, err := pgxutil.Select(ctx, db.q, `SELECT
    s.uuid, s.created_at, s.updated_at, s.deleted_at,
    s.title, s.artists, s.genre, s.edition, s.creator, s.language, s.year, s.comment, s.extra,
    s.gap, s.video_gap, s.start, s."end", s.preview_start, s.medley_start_beat, s.medley_end_beat, s.manual_medley,
    s.music_p1, s.music_p2, s.duet_singer1, s.duet_singer2,
    
    a.uuid, a.created_at, a.updated_at, a.deleted_at, a.type, a.size, a.checksum, a.duration,
    c.uuid, c.created_at, c.updated_at, c.deleted_at, c.type, c.size, c.checksum, c.width, c.height,
    v.uuid, v.created_at, v.updated_at, v.deleted_at, v.type, v.size, v.checksum, v.duration, v.width, v.height,
    b.uuid, b.created_at, b.updated_at, b.deleted_at, b.type, b.size, b.checksum, b.width, b.height
    FROM songs AS s
        LEFT OUTER JOIN files AS a ON s.audio_file_id = a.id
    	LEFT OUTER JOIN files AS c ON s.cover_file_id = c.id
    	LEFT OUTER JOIN files AS v ON s.video_file_id = v.id
    	LEFT OUTER JOIN files AS b ON s.background_file_id = b.id
	WHERE s.upload_id IS NULL
	LIMIT $1 OFFSET $2`, []any{limit, offset}, func(row pgx.CollectableRow) (*model.Song, error) {
		data, err := pgx.RowToStructByName[songRow](row)
		if err != nil {
			return nil, err
		}
		return data.ToModel(), nil
	})
	return songs, total, err
}

func (db *db) UpdateSong(ctx context.Context, song *model.Song) error {
	row, err := pgxutil.SelectRow(ctx, db.q, `UPDATE songs SET 
		title = $2, artists = $3, genre = $4, edition = $5, creator = $6, language = $7, year = $8, comment = $9, extra = $10,
		gap = $11, video_gap = $12, start = $13, "end" = $14, preview_start = $15, medley_start_beat = $16, medley_end_beat = $17, manual_medley = $18,
		music_p1 = $19, music_p2 = $20, duet_singer1 = $21, duet_singer2 = $22,
		audio_file_id = CASE WHEN $23 IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $23) END,
		cover_file_id = CASE WHEN $24 IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $24) END,
		video_file_id = CASE WHEN $25 IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $25) END,
		background_file_id = CASE WHEN $26 IS NULL THEN NULL ELSE (SELECT id FROM files WHERE uuid = $26) END
	WHERE uuid = $1
	RETURNING updated_at, audio_file_id, cover_file_id, video_file_id, background_file_id`, []any{
		song.UUID,
		song.Title, song.Artists, song.Genre, song.Edition, song.Creator, song.Language, song.Year, song.Comment, song.CustomTags,
		song.Gap, song.VideoGap, song.Start, song.End, song.PreviewStart, song.MedleyStartBeat, song.MedleyEndBeat, song.NoAutoMedley,
		(*dbutil.Music)(song.MusicP1), (*dbutil.Music)(song.MusicP2), song.DuetSinger1, song.DuetSinger2,
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

func (db *db) DeleteSongByUUID(ctx context.Context, id uuid.UUID) (bool, error) {
	_, err := pgxutil.ExecRow(ctx, db.q, `DELETE FROM songs WHERE uuid = $1`, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
