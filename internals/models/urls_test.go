package models

import (
	"testing"
)

func TestUrlExists(t *testing.T) {
	db := newTestDB(t)
	m := ShortenerDataModel{db}
	expected := ShortenerData{
		OriginalURL:  "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/424",
		ShortenedURL: "https://4C2P1PC8+",
		Clicks:       0,
	}
	data, err := m.Get("https://4C2P1PC8+")
	if *data != expected {
		t.Errorf("Expected %+v. Got: %+v", expected, data)
	}
	if err != nil {
		t.Errorf("Did not expect an error to be returned.")
	}
}

func TestCanInsertUrls(t *testing.T) {
	db := newTestDB(t)
	m := ShortenerDataModel{db}
	testData := ShortenerData{
		OriginalURL:  "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/404",
		ShortenedURL: "https://6C2P1PC8+",
		Clicks:       200,
	}
	affected, err := m.Insert(testData.OriginalURL, testData.ShortenedURL, testData.Clicks)
	if affected != 1 {
		t.Errorf("Expected %d, got %d", 1, affected)
	}
	if err != nil {
		t.Errorf("Did not expect an error to be returned.")
	}

	rows, _ := m.Latest()
	if len(rows) != 2 {
		t.Errorf("Incorrect number of rows returned. Expected %d; got %d", 2, len(rows))
	}
}

func TestCanRetrieveAllUrls(t *testing.T) {
	db := newTestDB(t)
	m := ShortenerDataModel{db}
	testData := ShortenerData{
		OriginalURL:  "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/424",
		ShortenedURL: "https://4C2P1PC8+",
		Clicks:       0,
	}
	rows, err := m.Latest()
	if err != nil {
		t.Errorf("Did not expect an error to be returned.")
	}
	if len(rows) != 1 {
		t.Errorf("Incorrect number of rows returned. Expected %d; got %d", 1, len(rows))
	}
	if *rows[0] != testData {
		t.Errorf("Incorrect URL data returned. Expected %+v, got %+v", testData, *rows[0])
	}
}

func TestCanIncrementUrlClicks(t *testing.T) {
	db := newTestDB(t)
	m := ShortenerDataModel{db}
	testData := ShortenerData{
		OriginalURL:  "https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/424",
		ShortenedURL: "https://4C2P1PC8+",
		Clicks:       1,
	}
	err := m.IncrementClicks(testData.ShortenedURL)
	if err != nil {
		t.Errorf("Did not expect an error to be returned.")
	}
	data, _ := m.Get("https://4C2P1PC8+")
	if data.Clicks != 1 {
		t.Errorf("Incorrect number of URL clicks returned. Expected %d. Got: %d", 1, data.Clicks)
	}
}
