package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
)

type MediaRepository interface {
	List() ([]model.MediaEntry, error)
	ListPaginated(lastDate *model.LocalDate, lastID int64, limit int) ([]model.MediaEntry, error)
	ListSince(date model.LocalDate) ([]model.MediaEntry, error)
	ListSincePaginated(targetDate model.LocalDate, lastDate *model.LocalDate, lastID int64, limit int) ([]model.MediaEntry, error)
	ListBetween(startDate model.LocalDate, finishDate model.LocalDate, lastDate *model.LocalDate, lastID int64, limit int) ([]model.MediaEntry, error)
	ListByDate(date model.LocalDate) ([]model.MediaEntry, error)

	Count() (int, error)
	CountSince(date model.LocalDate) (int, error)
	CountSearch(searchTerm string) (int, error)
	CountBetween(startDate model.LocalDate, finishDate model.LocalDate) (int, error)
	CountTimelineDays() (int, error)
	Exists(id int64) (bool, error)

	SaveBatch(entries []model.MediaEntry) error
	UpdateBatch(entries []model.MediaEntry) ([]model.MediaEntry, error)
	DeleteBatch(ids []int64) error
	Import(entries []model.MediaEntry) error

	Search(searchTerm string) ([]model.MediaEntry, error)
	SearchPaginated(searchTerm string, lastDate *model.LocalDate, lastID int64, limit int) ([]model.MediaEntry, error)
	GetStats(title string) (*model.TitleStats, bool, error)
	GetRatings(months int) ([]model.MediaRating, error)
	GetRatingsBetween(startDate model.LocalDate, finishDate model.LocalDate) ([]model.MediaRating, error)
	GetTimelinePaginated(lastDate *model.LocalDate, limit int) ([]model.TimelineItem, error)
	GetTimelineAll() ([]model.TimelineItem, error)
}

type postgresMediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) MediaRepository {
	return &postgresMediaRepository{db: db}
}

func (r *postgresMediaRepository) List() ([]model.MediaEntry, error) {
	query := `SELECT id, 
       		      title, 
       		      date_actual, 
       		      is_finished, 
       		      media_type, 
       		      media_genre,
       		      is_dropped,
       		      media_comment 
	          FROM titles 
	          ORDER BY date_actual DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := parseRows(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *postgresMediaRepository) ListPaginated(
	lastDate *model.LocalDate,
	lastID int64,
	limit int) ([]model.MediaEntry, error) {
	var query string
	var args []any

	if lastDate != nil && lastID > 0 {
		query = `SELECT 
                     id,
                     title,
                     date_actual,
                     is_finished,
                     media_type,
                     media_genre,
                     is_dropped,
                     media_comment
                 FROM titles
                 WHERE (date_actual, id) < ($1, $2)
                 ORDER BY date_actual DESC, id DESC 
                 LIMIT $3`
		args = append(args, *lastDate, lastID, limit)
	} else {
		query = `SELECT 
                     id,
                     title,
                     date_actual,
                     is_finished,
                     media_type,
                     media_genre,
                     is_dropped,
                     media_comment
                 FROM 
                     titles 
                 ORDER BY 
                     date_actual DESC, id DESC
                 LIMIT $1`
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseRows(rows)
}

func (r *postgresMediaRepository) ListSince(date model.LocalDate) ([]model.MediaEntry, error) {
	query := `SELECT id, 
       	          title, 
       		      date_actual, 
       		      is_finished, 
       		      media_type, 
       		      media_genre,
       		      is_dropped,
       		      media_comment 
	          FROM titles 
	          WHERE date_actual >= $1 
	          ORDER BY date_actual DESC`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := parseRows(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *postgresMediaRepository) ListSincePaginated(
	targetDate model.LocalDate,
	lastDate *model.LocalDate,
	lastID int64,
	limit int) ([]model.MediaEntry, error) {
	var query string
	var args []any

	if lastDate != nil && lastID > 0 {
		query = `SELECT id, 
       		          title, 
       		          date_actual, 
       		          is_finished, 
       		          media_type, 
       		          media_genre,
       		          is_dropped,
       		          media_comment 
				  FROM titles
				  WHERE date_actual >= $1
				      AND (date_actual, id) < ($2, $3)
				  ORDER BY 
				      date_actual DESC, 
				      id DESC
				  LIMIT $4`
		args = append(args, targetDate, *lastDate, lastID, limit)
	} else {
		query = `SELECT id, 
       		          title, 
       		          date_actual, 
       		          is_finished, 
       		          media_type, 
       		          media_genre,
       		          is_dropped,
       		          media_comment 
				  FROM titles
				  WHERE date_actual >= $1
				  ORDER BY 
				      date_actual DESC, 
				      id DESC
				  LIMIT $2`
		args = append(args, targetDate, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseRows(rows)
}

func (r *postgresMediaRepository) ListBetween(
	startDate model.LocalDate,
	finishDate model.LocalDate,
	lastDate *model.LocalDate,
	lastID int64,
	limit int) ([]model.MediaEntry, error) {
	var query string
	var args []any

	if lastDate != nil && lastID > 0 {
		query = `SELECT id, 
       		          title, 
       		          date_actual, 
       		          is_finished, 
       		          media_type, 
       		          media_genre,
       		          is_dropped,
       		          media_comment 
				  FROM titles
				  WHERE date_actual >= $1
				    AND date_actual <= $2
				    AND (date_actual, id) < ($3, $4)
				  ORDER BY 
				      date_actual DESC, 
				      id DESC
				  LIMIT $5`
		args = append(args, startDate, finishDate, *lastDate, lastID, limit)
	} else {
		query = `SELECT id, 
       		          title, 
       		          date_actual, 
       		          is_finished, 
       		          media_type, 
       		          media_genre,
       		          is_dropped,
       		          media_comment 
				  FROM titles
				  WHERE date_actual >= $1
				    AND date_actual <= $2
				  ORDER BY 
				      date_actual DESC, 
				      id DESC
				  LIMIT $3`
		args = append(args, startDate, finishDate, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseRows(rows)
}

func (r *postgresMediaRepository) ListByDate(date model.LocalDate) ([]model.MediaEntry, error) {
	query := `SELECT id, 
       		      title, 
       		      date_actual, 
       		      is_finished, 
       		      media_type, 
       		      media_genre,
       		      is_dropped,
       		      media_comment 
	          FROM titles 
	          WHERE date_actual = $1`
	rows, err := r.db.Query(query, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := parseRows(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *postgresMediaRepository) Count() (int, error) {
	query := `SELECT COUNT(*) FROM titles`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *postgresMediaRepository) CountSince(date model.LocalDate) (int, error) {
	query := `SELECT COUNT(*) FROM titles WHERE date_actual >= $1`
	var count int
	err := r.db.QueryRow(query, date).Scan(&count)
	return count, err
}

func (r *postgresMediaRepository) CountSearch(searchTerm string) (int, error) {
	query := `SELECT COUNT(*) 
			  FROM titles 
			  WHERE title ILIKE $1
			      OR media_comment ILIKE $1
			      OR media_type ILIKE $1`
	pattern := fmt.Sprintf("%%%s%%", searchTerm)
	var count int
	err := r.db.QueryRow(query, pattern).Scan(&count)
	return count, err
}

func (r *postgresMediaRepository) CountBetween(startDate model.LocalDate, finishDate model.LocalDate) (int, error) {
	query := `SELECT COUNT(*) FROM titles WHERE date_actual >= $1 AND date_actual <= $2`
	var count int
	err := r.db.QueryRow(query, startDate, finishDate).Scan(&count)
	return count, err
}

func (r *postgresMediaRepository) CountTimelineDays() (int, error) {
	query := `SELECT COALESCE(CURRENT_DATE - MIN(date_actual) + 1, 1) FROM titles`
	var count int
	err := r.db.QueryRow(query).Scan(&count)
	return count, err
}

func (r *postgresMediaRepository) Exists(id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM titles WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (r *postgresMediaRepository) SaveBatch(entries []model.MediaEntry) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO titles (title,
                    date_actual, 
                    is_finished, 
                    media_type, 
                    media_genre,
                    is_dropped,
                    media_comment) 
		          VALUES ($1, $2, $3, $4, $5, $6, $7)
		          RETURNING id
		          `)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i := range entries {
		err := stmt.QueryRow(entries[i].Title,
			entries[i].Date,
			entries[i].IsFinished,
			entries[i].Type,
			entries[i].Genre,
			entries[i].IsDropped,
			entries[i].Comment,
		).Scan(&entries[i].ID)

		if err != nil {
			return fmt.Errorf("failed to insert entry at index %d: %w", i, err)
		}
	}

	return tx.Commit()
}

func (r *postgresMediaRepository) UpdateBatch(entries []model.MediaEntry) ([]model.MediaEntry, error) {
	if len(entries) == 0 {
		return nil, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE titles SET 
                      title=$1, 
                      date_actual=$2, 
                      is_finished=$3,
                      media_type=COALESCE($4, media_type),
                      media_genre=COALESCE($5, media_genre),
                      is_dropped=$6,
                      media_comment=COALESCE($7, media_comment)
                  WHERE id = $8
                  RETURNING 
                      id, 
                      title, 
                      date_actual,
                      is_finished,
                      media_type,
                      media_genre,
                      is_dropped,
                      media_comment
                  `)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	updatedEntries := make([]model.MediaEntry, 0, len(entries))

	for i, entry := range entries {
		var updated model.MediaEntry

		err := stmt.QueryRow(
			entry.Title,
			entry.Date,
			entry.IsFinished,
			entry.Type,
			entry.Genre,
			entry.IsDropped,
			entry.Comment,
			entry.ID,
		).Scan(
			&updated.ID,
			&updated.Title,
			&updated.Date,
			&updated.IsFinished,
			&updated.Type,
			&updated.Genre,
			&updated.IsDropped,
			&updated.Comment,
		)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("media entry does not exist with id: %d", entry.ID)
			}
			return nil, fmt.Errorf("failed to update entry at index %d (ID %d): %w", i, entry.ID, err)
		}

		updatedEntries = append(updatedEntries, updated)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return updatedEntries, nil
}

func (r *postgresMediaRepository) DeleteBatch(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`DELETE FROM titles WHERE id = $1`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.Exec(id); err != nil {
			return fmt.Errorf("failed to delete entry with id %d: %w", id, err)
		}
	}

	return tx.Commit()
}

func (r *postgresMediaRepository) Import(entries []model.MediaEntry) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("error beginning transaction: %s", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("TRUNCATE TABLE titles RESTART IDENTITY"); err != nil {
		return fmt.Errorf("error truncating table `titles`: %s", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO titles (title, 
                    date_actual, 
                    is_finished, 
                    media_type, 
                    media_genre, 
                    is_dropped, 
                    media_comment)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`)
	if err != nil {
		return fmt.Errorf("error preparing statement: %s", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.Exec(entry.Title,
			entry.Date,
			entry.IsFinished,
			entry.Type,
			entry.Genre,
			entry.IsDropped,
			entry.Comment)
		if err != nil {
			return fmt.Errorf("error inserting entry '%v - %s': %s", entry.Date, entry.Title, err)
		}
	}

	return tx.Commit()
}

func (r *postgresMediaRepository) Search(searchTerm string) ([]model.MediaEntry, error) {
	query := `SELECT id, 
       		      title, 
       		      date_actual, 
       		      is_finished, 
       		      media_type, 
       		      media_genre, 
       		      is_dropped, 
       		      media_comment
			  FROM titles
			  WHERE title ILIKE $1
			  OR media_comment ILIKE $1
			  OR media_type ILIKE $1
			  ORDER BY date_actual DESC`

	pattern := fmt.Sprintf("%%%s%%", searchTerm)
	rows, err := r.db.Query(query, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result, err := parseRows(rows)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (r *postgresMediaRepository) SearchPaginated(searchTerm string, lastDate *model.LocalDate, lastID int64, limit int) ([]model.MediaEntry, error) {
	pattern := fmt.Sprintf("%%%s%%", searchTerm)
	var query string
	var args []any

	if lastDate != nil && lastID > 0 {
		query = `SELECT 
                     id,
                     title,
                     date_actual,
                     is_finished,
                     media_type,
                     media_genre,
                     is_dropped,
                     media_comment
                 FROM 
                     titles 
                 WHERE 
                     (title ILIKE $1 
                          OR media_comment ILIKE $1
                          OR media_type ILIKE $1)
                     AND (date_actual, id) < ($2, $3)
                 ORDER BY
                     date_actual DESC,
                     id DESC
                 LIMIT $4`
		args = append(args, pattern, *lastDate, lastID, limit)
	} else {
		query = `SELECT 
                     id,
                     title,
                     date_actual,
                     is_finished,
                     media_type,
                     media_genre,
                     is_dropped,
                     media_comment
                 FROM 
                     titles 
                 WHERE 
                     title ILIKE $1 
                          OR media_comment ILIKE $1
                          OR media_type ILIKE $1
                 ORDER BY
                     date_actual DESC,
                     id DESC
                 LIMIT $2`
		args = append(args, pattern, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return parseRows(rows)
}

func (r *postgresMediaRepository) GetStats(title string) (*model.TitleStats, bool, error) {
	query := `
		SELECT
		    EXISTS(SELECT 1 FROM titles WHERE title = $1), 
			COUNT(*) FILTER	(WHERE date_actual > NOW() - INTERVAL '3 days'),
			COUNT(*) FILTER	(WHERE date_actual > NOW() - INTERVAL '7 days'),
			COUNT(*) FILTER	(WHERE date_actual > NOW() - INTERVAL '30 days'),
			COUNT(*) FILTER	(WHERE date_actual > NOW() - INTERVAL '180 days'),
			COUNT(*) 
		FROM titles
		WHERE title = $1`

	var stats model.TitleStats
	stats.Title = title
	var exists bool

	err := r.db.QueryRow(query, title).Scan(
		&exists,
		&stats.Last3days,
		&stats.Last7days,
		&stats.Last30days,
		&stats.Last180days,
		&stats.Total,
	)
	if err != nil {
		return nil, false, err
	}

	return &stats, exists, nil
}

func (r *postgresMediaRepository) GetRatings(months int) ([]model.MediaRating, error) {
	query := `SELECT 
    				t.title,
					t.media_type,
					COUNT(t.id) AS total,
					COUNT(t.id) FILTER (WHERE is_finished) AS finished
	          FROM 
					titles AS t 
	          WHERE
					($1 <= 0 OR t.date_actual > CURRENT_DATE - ($1 * INTERVAL '1 month'))
	          GROUP BY 
					t.title,
					t.media_type
			  HAVING
					COUNT(t.id) < 999
					AND COUNT(t.id) >= 1
			  ORDER BY 
					finished DESC,
					total DESC,
					t.media_type;`

	rows, err := r.db.Query(query, months)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var result []model.MediaRating
	for rows.Next() {
		var m model.MediaRating
		err := rows.Scan(&m.Title, &m.Type, &m.Total, &m.Finished)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *postgresMediaRepository) GetRatingsBetween(startDate model.LocalDate, finishDate model.LocalDate) ([]model.MediaRating, error) {
	query := `SELECT 
    				t.title,
					t.media_type,
					COUNT(t.id) AS total,
					COUNT(t.id) FILTER (WHERE is_finished) AS finished
	          FROM 
					titles AS t 
	          WHERE
					t.date_actual >= $1
					AND t.date_actual <= $2
	          GROUP BY 
					t.title,
					t.media_type
			  HAVING
					COUNT(t.id) < 999
					AND COUNT(t.id) >= 1
			  ORDER BY 
					finished DESC,
					total DESC,
					t.media_type;`
	rows, err := r.db.Query(query, startDate, finishDate)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {

		}
	}(rows)

	var result []model.MediaRating
	for rows.Next() {
		var m model.MediaRating
		err := rows.Scan(&m.Title, &m.Type, &m.Total, &m.Finished)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *postgresMediaRepository) GetTimelinePaginated(lastDate *model.LocalDate, limit int) ([]model.TimelineItem, error) {
	var query string
	var args []any

	if lastDate != nil {
		query = `WITH date_range AS (
						SELECT 
							COALESCE(MIN(date_actual), CURRENT_DATE) AS min_date,
							CURRENT_DATE AS max_date
						FROM titles
				 )
				 SELECT
						d.day::date AS timeline_date,
						EXISTS(SELECT 1 FROM titles AS t WHERE t.date_actual = d.day::date) AS has_media
				 FROM date_range dr,
				 GENERATE_SERIES(dr.min_date, dr.max_date, '1 day'::interval) AS d(day)
				 WHERE d.day::date < $1
				 ORDER BY timeline_date DESC
				 LIMIT $2`
		args = append(args, *lastDate, limit)
	} else {
		query = `WITH date_range AS (
						SELECT 
							COALESCE(MIN(date_actual), CURRENT_DATE) AS min_date,
							CURRENT_DATE AS max_date
						FROM titles
						)
					SELECT
						d.day::date AS timeline_date,
						EXISTS(SELECT 1 FROM titles AS t WHERE t.date_actual = d.day::date) AS has_media
					FROM date_range dr,
					GENERATE_SERIES(dr.min_date, dr.max_date, '1 day'::interval) AS d(day)
					ORDER BY timeline_date DESC
					LIMIT $1`
		args = append(args, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.TimelineItem
	for rows.Next() {
		var item model.TimelineItem
		if err := rows.Scan(&item.Date, &item.HasMedia); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *postgresMediaRepository) GetTimelineAll() ([]model.TimelineItem, error) {
	query := `WITH date_range AS (
						SELECT 
							COALESCE(MIN(date_actual), CURRENT_DATE) AS min_date,
							CURRENT_DATE AS max_date
						FROM titles
				 )
				 SELECT
						d.day::date AS timeline_date,
						EXISTS(SELECT 1 FROM titles AS t WHERE t.date_actual = d.day::date) AS has_media
				 FROM date_range dr,
				 GENERATE_SERIES(dr.min_date, dr.max_date, '1 day'::interval) AS d(day)
				 ORDER BY timeline_date DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.TimelineItem
	for rows.Next() {
		var item model.TimelineItem
		if err := rows.Scan(&item.Date, &item.HasMedia); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func parseRows(rows *sql.Rows) ([]model.MediaEntry, error) {
	var result []model.MediaEntry
	for rows.Next() {
		var m model.MediaEntry
		err := rows.Scan(&m.ID, &m.Title, &m.Date, &m.IsFinished, &m.Type, &m.Genre, &m.IsDropped, &m.Comment)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
