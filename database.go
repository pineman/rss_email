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

func Close() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
