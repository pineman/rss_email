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

	emailSender = NewSender(cfg.GmailAddress, cfg.GmailAppPassword, cfg.RecipientEmail)

	checkFeeds()

	stopChan := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
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

	log.Println("Scheduler started - checking feeds every 30 minutes")

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
		feedName, items, err := FetchFeed(feedURL)
		if err != nil {
			log.Printf("Error fetching feed %s: %v", feedURL, err)
			continue
		}

		if len(items) == 0 {
			continue
		}

		hasFeedItems, err := HasFeedItems(feedURL)
		if err != nil {
			log.Printf("Error checking feed items for %s: %v", feedURL, err)
			continue
		}

		if hasFeedItems {
			processExistingFeed(feedURL, feedName, items)
		} else {
			processNewFeed(feedURL, feedName, items)
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
