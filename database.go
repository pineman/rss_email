package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func Initialize(dbPath string) error {
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS sent_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_url TEXT NOT NULL,
		item_guid TEXT NOT NULL,
		sent_at TIMESTAMP NOT NULL,
		UNIQUE(feed_url, item_guid)
	);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	createFeedMetadataTableSQL := `
	CREATE TABLE IF NOT EXISTS feed_metadata (
		feed_url TEXT PRIMARY KEY,
		last_modified TEXT,
		etag TEXT,
		last_checked TIMESTAMP NOT NULL,
		last_poll_status INTEGER,
		next_check_after TIMESTAMP,
		error_count INTEGER DEFAULT 0
	);
	`

	if _, err := db.Exec(createFeedMetadataTableSQL); err != nil {
		return fmt.Errorf("failed to create feed_metadata table: %w", err)
	}

	// Schema migration for existing tables
	// We ignore errors here as columns might already exist
	_, _ = db.Exec("ALTER TABLE feed_metadata ADD COLUMN next_check_after TIMESTAMP")
	_, _ = db.Exec("ALTER TABLE feed_metadata ADD COLUMN error_count INTEGER DEFAULT 0")

	createIndexSQL := `
	CREATE INDEX IF NOT EXISTS idx_feed_guid 
	ON sent_items(feed_url, item_guid);
	`

	if _, err := db.Exec(createIndexSQL); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

func IsItemSent(feedURL, itemGUID string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM sent_items WHERE feed_url = ? AND item_guid = ?"
	err := db.QueryRow(query, feedURL, itemGUID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if item is sent: %w", err)
	}
	return count > 0, nil
}

func MarkItemSent(feedURL, itemGUID string) error {
	query := "INSERT INTO sent_items (feed_url, item_guid, sent_at) VALUES (?, ?, ?)"
	_, err := db.Exec(query, feedURL, itemGUID, time.Now())
	if err != nil {
		if err.Error() == "UNIQUE constraint failed: sent_items.feed_url, sent_items.item_guid" {
			return nil
		}
		return fmt.Errorf("failed to mark item as sent: %w", err)
	}
	return nil
}

func HasFeedItems(feedURL string) (bool, error) {
	var count int
	query := "SELECT COUNT(*) FROM sent_items WHERE feed_url = ?"
	err := db.QueryRow(query, feedURL).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if feed has items: %w", err)
	}
	return count > 0, nil
}

func GetSentCount() (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM sent_items"
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get sent count: %w", err)
	}
	return count, nil
}

type FeedMetadata struct {
	FeedURL        string
	LastModified   string
	ETag           string
	LastChecked    time.Time
	LastPollStatus int
	NextCheckAfter *time.Time
	ErrorCount     int
}

func GetFeedMetadata(feedURL string) (*FeedMetadata, error) {
	var metadata FeedMetadata
	query := `SELECT feed_url, COALESCE(last_modified, ''), COALESCE(etag, ''), 
	          last_checked, COALESCE(last_poll_status, 0), next_check_after, COALESCE(error_count, 0)
	          FROM feed_metadata WHERE feed_url = ?`

	err := db.QueryRow(query, feedURL).Scan(
		&metadata.FeedURL,
		&metadata.LastModified,
		&metadata.ETag,
		&metadata.LastChecked,
		&metadata.LastPollStatus,
		&metadata.NextCheckAfter,
		&metadata.ErrorCount,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feed metadata: %w", err)
	}

	return &metadata, nil
}

func UpdateFeedError(feedURL string, pollStatus int, errorCount int, nextCheckAfter time.Time) error {
	query := `UPDATE feed_metadata 
	          SET last_checked = ?, last_poll_status = ?, error_count = ?, next_check_after = ?
	          WHERE feed_url = ?`

	result, err := db.Exec(query, time.Now(), pollStatus, errorCount, nextCheckAfter, feedURL)
	if err != nil {
		return fmt.Errorf("failed to update feed status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		// New feed failed on first try
		insertQuery := `INSERT INTO feed_metadata (feed_url, last_modified, etag, last_checked, last_poll_status, error_count, next_check_after)
		                VALUES (?, '', '', ?, ?, ?, ?)`
		_, err := db.Exec(insertQuery, feedURL, time.Now(), pollStatus, errorCount, nextCheckAfter)
		if err != nil {
			return fmt.Errorf("failed to insert new feed status: %w", err)
		}
	}

	return nil
}

func UpdateFeedSuccess(feedURL, lastModified, etag string, pollStatus int, nextCheckAfter time.Time) error {
	query := `INSERT INTO feed_metadata (feed_url, last_modified, etag, last_checked, last_poll_status, error_count, next_check_after)
	          VALUES (?, ?, ?, ?, ?, 0, ?)
	          ON CONFLICT(feed_url) DO UPDATE SET
	            last_modified = excluded.last_modified,
	            etag = excluded.etag,
	            last_checked = excluded.last_checked,
	            last_poll_status = excluded.last_poll_status,
				error_count = 0,
				next_check_after = excluded.next_check_after`

	_, err := db.Exec(query, feedURL, lastModified, etag, time.Now(), pollStatus, nextCheckAfter)
	if err != nil {
		return fmt.Errorf("failed to update feed metadata: %w", err)
	}

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
