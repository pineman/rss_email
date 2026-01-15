package main

import (
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

var (
	cfg         *Config
	emailSender *Sender
)

const StandardInterval = 60 * time.Minute

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
		ticker := time.NewTicker(StandardInterval)
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

	log.Printf("Scheduler started - checking feeds every %v", StandardInterval)

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

		// FRB037 & Backoff: Check if it's time to poll
		if metadata != nil {
			if metadata.NextCheckAfter != nil && time.Now().Before(*metadata.NextCheckAfter) ||
				time.Since(metadata.LastChecked) < StandardInterval {
				log.Printf("Skipping %s, next check after %v", feedURL, metadata.NextCheckAfter)
				continue
			}
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

			status := 0
			retryAfter := ""
			if result != nil {
				status = result.StatusCode
				retryAfter = result.RetryAfter
			}

			currentErrorCount := 0
			if metadata != nil {
				currentErrorCount = metadata.ErrorCount
			}
			newErrorCount := currentErrorCount + 1

			nextCheck := calculateBackoff(status, retryAfter, newErrorCount)

			// FRB016: Only update status/error/schedule, keep old cache headers
			if err := UpdateFeedError(feedURL, status, newErrorCount, nextCheck); err != nil {
				log.Printf("Error updating status for %s: %v", feedURL, err)
			}
			continue
		}

		// Success or 304
		// Standard interval
		nextCheck := time.Now().Add(StandardInterval)
		if err := UpdateFeedSuccess(feedURL, result.LastModified, result.ETag, result.StatusCode, nextCheck); err != nil {
			log.Printf("Error updating metadata for %s: %v", feedURL, err)
		}

		if result.NotModified || len(result.Items) == 0 {
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

func calculateBackoff(status int, retryAfter string, errorCount int) time.Time {
	// FRB114: Stop on 410 Gone
	if status == http.StatusGone {
		log.Printf("Feed returned 410 Gone - disabling feed (FRB114)")
		// Disable for a year (essentially forever)
		return time.Now().Add(365 * 24 * time.Hour)
	}

	// FRB020: Check Retry-After
	if retryDuration := parseRetryAfter(retryAfter); retryDuration > 0 {
		return time.Now().Add(retryDuration)
	}

	// FRB110-119: Exponential backoff for errors
	// Base interval StandardInterval.
	// 1 error: 60m
	// 2 errors: 120m
	// 3 errors: 240m
	// ...
	// Cap at 24 hours
	backoff := StandardInterval * time.Duration(math.Pow(2, float64(errorCount-1)))
	if backoff > 24*time.Hour {
		backoff = 24 * time.Hour
	}

	// Ensure at least base interval
	if backoff < StandardInterval {
		backoff = StandardInterval
	}

	return time.Now().Add(backoff)
}

func parseRetryAfter(header string) time.Duration {
	if header == "" {
		return 0
	}
	// Try parsing as seconds
	if seconds, err := strconv.Atoi(header); err == nil {
		return time.Duration(seconds) * time.Second
	}
	// Try parsing as HTTP date
	if t, err := http.ParseTime(header); err == nil {
		return time.Until(t)
	}
	return 0
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
