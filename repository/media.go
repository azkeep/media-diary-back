package repository

import (
	"database/sql"
	"fmt"
	"github.com/azkeep/MediaDiary/backend-go/model"
)

type MediaRepository interface {
	FindAllByOrderByDateDesc() ([]model.MediaEntry, error)
	FindAllByDateGreaterThanEqualOrderByDateDesc(date model.LocalDate) ([]model.MediaEntry, error)
	FindByDate(date model.LocalDate) ([]model.MediaEntry, error)
	ExistsByID(id int64) (bool, error)
	Save(media *model.MediaEntry) error
	DeleteByID(id int64) error
	GetTitleStats(title string) (*model.StatsResponse, bool, error)
	FindAllUniqueTitlesByRating(months int) ([]model.MediaRating, error)
	ImportBatch(entries []model.MediaEntry) error
	SearchEntries(searchTerm string) ([]model.MediaEntry, error)
}

type postgresMediaRepository struct {
	db *sql.DB
}

func NewMediaRepository(db *sql.DB) MediaRepository {
	return &postgresMediaRepository{db: db}
}

func (r *postgresMediaRepository) FindAllByOrderByDateDesc() ([]model.MediaEntry, error) {
	query := `SELECT id, 
       		      title, 
       		      date_actual, 
       		      is_finished, 
       		      media_type, 
       		      media_genre, 
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

func (r *postgresMediaRepository) FindAllByDateGreaterThanEqualOrderByDateDesc(date model.LocalDate) ([]model.MediaEntry, error) {
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

func (r *postgresMediaRepository) FindByDate(date model.LocalDate) ([]model.MediaEntry, error) {
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

func (r *postgresMediaRepository) ExistsByID(id int64) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM titles WHERE id = $1)`
	var exists bool
	err := r.db.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func (r *postgresMediaRepository) Save(media *model.MediaEntry) error {
	if media.ID == 0 {
		query := `INSERT INTO titles (title, 
                    date_actual, 
                    is_finished, 
                    media_type, 
                    media_genre,
                    is_dropped,
                    media_comment) 
		          VALUES ($1, $2, $3, $4, $5, $6, $7) 
		          RETURNING id`
		return r.db.QueryRow(query, media.Title, media.Date, media.IsFinished, media.Type, media.Genre, media.IsDropped, media.Comment).Scan(&media.ID)
	}

	query := `UPDATE titles 
	          SET title = $1, date_actual = $2, is_finished = $3, media_type = $4, media_genre = $5, media_comment = $6 
	          WHERE id = $7`
	_, err := r.db.Exec(query, media.Title, media.Date, media.IsFinished, media.Type, media.Genre, media.Comment, media.ID)
	return err
}

func (r *postgresMediaRepository) DeleteByID(id int64) error {
	query := `DELETE FROM titles WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *postgresMediaRepository) GetTitleStats(title string) (*model.StatsResponse, bool, error) {
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

	var stats model.StatsResponse
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

func (r *postgresMediaRepository) FindAllUniqueTitlesByRating(months int) ([]model.MediaRating, error) {
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

func (r *postgresMediaRepository) ImportBatch(entries []model.MediaEntry) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("Error beginning transaction: %s", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("TRUNCATE TABLE titles RESTART IDENTITY"); err != nil {
		return fmt.Errorf("Error truncating table `titles`: %s", err)
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
		return fmt.Errorf("Error preparing statement: %s", err)
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
			return fmt.Errorf("error inserting entry '%s - %s': %s", entry.Date, entry.Title, err)
		}
	}

	return tx.Commit()
}

func (r *postgresMediaRepository) SearchEntries(searchTerm string) ([]model.MediaEntry, error) {
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
