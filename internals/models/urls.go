package models

import (
	"database/sql"
)

// Stores an original URL, shortened URL, and the number of times the shortened URL was clicked
type ShortenerData struct {
	OriginalURL, ShortenedURL string
	Clicks                    int
}

type ShortenerDataModel struct {
	DB *sql.DB
}

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

func (m *ShortenerDataModel) Get(shortened string) (*ShortenerData, error) {
	return nil, nil
}

func (m *ShortenerDataModel) Latest() ([]*ShortenerData, error) {
	stmt := `SELECT original_url, shortened_url, clicks FROM urls`
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
