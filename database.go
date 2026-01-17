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

	schema := `
	CREATE TABLE IF NOT EXISTS sent_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		feed_url TEXT NOT NULL,
		item_guid TEXT NOT NULL,
		sent_at TIMESTAMP NOT NULL,
		UNIQUE(feed_url, item_guid)
	);
	CREATE TABLE IF NOT EXISTS feed_metadata (
		feed_url TEXT PRIMARY KEY,
		last_modified TEXT,
		etag TEXT,
		last_checked TIMESTAMP NOT NULL,
		last_poll_status INTEGER,
		next_check_after TIMESTAMP,
		error_count INTEGER DEFAULT 0
	);
	CREATE INDEX IF NOT EXISTS idx_feed_guid ON sent_items(feed_url, item_guid);
	`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
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
	query := "INSERT OR IGNORE INTO sent_items (feed_url, item_guid, sent_at) VALUES (?, ?, ?)"
	_, err := db.Exec(query, feedURL, itemGUID, time.Now())
	if err != nil {
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

type FeedMetadata struct {
	LastModified   string
	ETag           string
	LastChecked    time.Time
	LastPollStatus int
	NextCheckAfter *time.Time
	ErrorCount     int
}

func GetFeedMetadata(feedURL string) (*FeedMetadata, error) {
	var metadata FeedMetadata
	query := `SELECT COALESCE(last_modified, ''), COALESCE(etag, ''), 
	          last_checked, COALESCE(last_poll_status, 0), next_check_after, COALESCE(error_count, 0)
	          FROM feed_metadata WHERE feed_url = ?`

	err := db.QueryRow(query, feedURL).Scan(
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

func UpsertFeedMetadata(feedURL, lastModified, etag string, pollStatus, errorCount int, nextCheckAfter time.Time) error {
	query := `INSERT INTO feed_metadata (feed_url, last_modified, etag, last_checked, last_poll_status, error_count, next_check_after)
	          VALUES (?, ?, ?, ?, ?, ?, ?)
	          ON CONFLICT(feed_url) DO UPDATE SET
	            last_modified = excluded.last_modified,
	            etag = excluded.etag,
	            last_checked = excluded.last_checked,
	            last_poll_status = excluded.last_poll_status,
	            error_count = excluded.error_count,
	            next_check_after = excluded.next_check_after`

	_, err := db.Exec(query, feedURL, lastModified, etag, time.Now(), pollStatus, errorCount, nextCheckAfter)
	if err != nil {
		return fmt.Errorf("failed to upsert feed metadata: %w", err)
	}

	return nil
}

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
