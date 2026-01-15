package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	cfg         *Config
	emailSender *Sender
)

func main() {
	var err error
	cfg, err = Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := Initialize("data/rss_email.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer Close()

	emailSender = NewSender(cfg.GmailAppPassword)

	checkFeeds()

	stopChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(60 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				checkFeeds()
			case <-stopChan:
				return
			}
		}
	}()

	log.Println("Scheduler started - checking feeds every 60 minutes")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	close(stopChan)
	time.Sleep(100 * time.Millisecond)
}

func checkFeeds() {
	log.Println("Checking feeds...")

	for _, feedURL := range cfg.Feeds {
		metadata, err := GetFeedMetadata(feedURL)
		if err != nil {
			log.Printf("Error getting metadata for %s: %v", feedURL, err)
			continue
		}

		lastModified := ""
		etag := ""
		if metadata != nil {
			lastModified = metadata.LastModified
			etag = metadata.ETag
		}

		result, err := FetchFeed(feedURL, lastModified, etag)
		if err != nil {
			log.Printf("Error fetching feed %s: %v", feedURL, err)

			// Still update metadata even on error to track status
			if result != nil {
				if err := UpdateFeedMetadata(feedURL, result.LastModified, result.ETag, result.StatusCode); err != nil {
					log.Printf("Error updating metadata for %s: %v", feedURL, err)
				}
			}
			continue
		}

		if err := UpdateFeedMetadata(feedURL, result.LastModified, result.ETag, result.StatusCode); err != nil {
			log.Printf("Error updating metadata for %s: %v", feedURL, err)
		}

		if result.NotModified {
			log.Printf("Feed not modified: %s", feedURL)
			continue
		}

		if result.RateLimited {
			log.Printf("Rate limited for feed: %s - will retry later", feedURL)
			continue
		}

		if len(result.Items) == 0 {
			continue
		}

		hasFeedItems, err := HasFeedItems(feedURL)
		if err != nil {
			log.Printf("Error checking feed items for %s: %v", feedURL, err)
			continue
		}

		if hasFeedItems {
			processExistingFeed(feedURL, result.FeedTitle, result.Items)
		} else {
			processNewFeed(feedURL, result.FeedTitle, result.Items)
		}
	}

	log.Println("Done checking feeds.")
}

func processNewFeed(feedURL, feedName string, items []FeedItem) {
	mostRecent := GetMostRecentItem(items)
	if mostRecent == nil {
		return
	}

	sendItem(feedURL, feedName, *mostRecent)

	for _, item := range items {
		if item.GUID != mostRecent.GUID {
			if err := MarkItemSent(feedURL, item.GUID); err != nil {
				log.Printf("Error marking item as sent: %v", err)
			}
		}
	}
}

func processExistingFeed(feedURL, feedName string, items []FeedItem) {
	for _, item := range items {
		isSent, err := IsItemSent(feedURL, item.GUID)
		if err != nil {
			log.Printf("Error checking if item is sent: %v", err)
			continue
		}

		if !isSent {
			sendItem(feedURL, feedName, item)
		}
	}
}

func sendItem(feedURL, feedName string, item FeedItem) {
	subject, textBody, htmlBody := FormatRSSEmail(feedName, item)

	if err := emailSender.SendEmail(subject, textBody, htmlBody); err != nil {
		log.Printf("Error sending email for %s: %v", item.Title, err)
		return
	}

	if err := MarkItemSent(feedURL, item.GUID); err != nil {
		log.Printf("Error marking item as sent: %v", err)
		return
	}

	log.Printf("Sent: %s for %s", item.Title, feedURL)

	time.Sleep(1 * time.Second) // Rate limiting
}
