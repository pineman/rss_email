package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title       string
	Link        string
	GUID        string
	Published   string
	PublishedDT *time.Time
	Summary     string
}

func FetchFeed(url string) (string, []FeedItem, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	fp := gofeed.NewParser()
	fp.Client = client

	feed, err := fp.ParseURL(url)
	if err != nil {
		return url, nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	feedTitle := url
	if feed.Title != "" {
		feedTitle = feed.Title
	}

	items := make([]FeedItem, 0, len(feed.Items))
	for _, entry := range feed.Items {
		item := normalizeFeedItem(entry)
		if item != nil {
			items = append(items, *item)
		} else {
			log.Printf("Warning: Skipping item with no GUID or link - Title: %s, Feed: %s",
				entry.Title, url)
		}
	}

	return feedTitle, items, nil
}

func normalizeFeedItem(entry *gofeed.Item) *FeedItem {
	title := entry.Title
	if title == "" {
		title = "No Title"
	}

	link := entry.Link
	guid := entry.GUID
	if guid == "" {
		guid = link
	}

	if guid == "" && link == "" {
		return nil
	}

	published := "Unknown"
	var publishedDT *time.Time

	if entry.PublishedParsed != nil {
		publishedDT = entry.PublishedParsed
		published = entry.PublishedParsed.Format("2006-01-02 15:04:05")
	} else if entry.UpdatedParsed != nil {
		publishedDT = entry.UpdatedParsed
		published = entry.UpdatedParsed.Format("2006-01-02 15:04:05")
	} else if entry.Published != "" {
		published = entry.Published
	} else if entry.Updated != "" {
		published = entry.Updated
	}

	summary := getSummary(entry)

	return &FeedItem{
		Title:       title,
		Link:        link,
		GUID:        guid,
		Published:   published,
		PublishedDT: publishedDT,
		Summary:     summary,
	}
}

func getSummary(entry *gofeed.Item) string {
	if entry.Description != "" {
		return entry.Description
	}

	if entry.Content != "" {
		return entry.Content
	}

	return "No summary available."
}

func GetMostRecentItem(items []FeedItem) *FeedItem {
	if len(items) == 0 {
		return nil
	}

	var mostRecent *FeedItem
	for i := range items {
		if items[i].PublishedDT != nil {
			if mostRecent == nil || items[i].PublishedDT.After(*mostRecent.PublishedDT) {
				mostRecent = &items[i]
			}
		}
	}

	if mostRecent != nil {
		return mostRecent
	}

	return &items[0]
}
