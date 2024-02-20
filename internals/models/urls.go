package models

import (
	"database/sql"
	"errors"
)

type ShortenerDataInterface interface {
	Get(shortened string) (*ShortenerData, error)
	IncrementClicks(shortened string) error
	Insert(original string, shortened string, clicks int) (int, error)
	Latest() ([]*ShortenerData, error)
}

// Stores an original URL, shortened URL, and the number of times the shortened URL was clicked
type ShortenerData struct {
	OriginalURL, ShortenedURL string
	Clicks                    int
}

// ShortenerDataModel manages database interaction for the URL shortener data
type ShortenerDataModel struct {
	DB *sql.DB
}

// Insert inserts a new record into the urls table
func (m *ShortenerDataModel) Insert(original string, shortened string, clicks int) (int, error) {
	stmt := `INSERT INTO urls  (original_url, shortened_url, clicks) VALUES(?, ?, ?)`
	result, err := m.DB.Exec(stmt, original, shortened, clicks)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rowsAffected), nil
}

// Get retrieves a record from the urls table identifying that record by the shortened URL
func (m *ShortenerDataModel) Get(shortened string) (*ShortenerData, error) {
	stmt := `SELECT original_url, shortened_url, clicks FROM urls WHERE shortened_url = ?`
	row := m.DB.QueryRow(stmt, shortened)
	data := &ShortenerData{}
	err := row.Scan(&data.OriginalURL, &data.ShortenedURL, &data.Clicks)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var ErrNoRecord = errors.New("models: no matching record found")
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return data, nil
}

// IncrementClicks increments the number of clicks for a shortened URL by one
func (m *ShortenerDataModel) IncrementClicks(shortened string) error {
	stmt := `UPDATE urls SET clicks = clicks + 1 WHERE shortened_url = ?`
	_, err := m.DB.Exec(stmt, shortened)
	if err != nil {
		return err
	}

	return nil
}

// Latest retrieves all of the records from the urls table in the database
func (m *ShortenerDataModel) Latest() ([]*ShortenerData, error) {
	stmt := `SELECT original_url, shortened_url, clicks FROM urls ORDER BY created DESC, original_url ASC`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*ShortenerData{}
	for rows.Next() {
		url := &ShortenerData{}
		err := rows.Scan(&url.OriginalURL, &url.ShortenedURL, &url.Clicks)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
