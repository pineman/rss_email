package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title     string
	Link      string
	GUID      string
	Published *time.Time
	Summary   string
}

type FeedResult struct {
	FeedTitle    string
	Items        []FeedItem
	LastModified string
	ETag         string
	RetryAfter   string
	StatusCode   int
	NotModified  bool
}

func FetchFeed(url string, lastModified, etag string) (*FeedResult, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set conditional request headers if we have them
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	req.Header.Set("User-Agent", "rss_email/1.0 (+https://github.com/pineman/rss_email)")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	result := &FeedResult{
		StatusCode:   resp.StatusCode,
		LastModified: resp.Header.Get("Last-Modified"),
		ETag:         resp.Header.Get("ETag"),
		RetryAfter:   resp.Header.Get("Retry-After"),
	}

	if resp.StatusCode == http.StatusNotModified {
		result.NotModified = true
		result.FeedTitle = url
		return result, nil
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return result, fmt.Errorf("rate limited (429), Retry-After: %s", result.RetryAfter)
	}

	if resp.StatusCode != http.StatusOK {
		return result, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("failed to read response body: %w", err)
	}

	// Fix common malformed HTML in feeds (e.g., </br> instead of <br/>)
	content := strings.ReplaceAll(string(body), "</br>", "<br/>")

	fp := gofeed.NewParser()
	feed, err := fp.ParseString(content)
	if err != nil {
		return result, fmt.Errorf("failed to parse feed: %w", err)
	}

	feedTitle := url
	if feed.Title != "" {
		feedTitle = feed.Title
	}
	result.FeedTitle = feedTitle

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
	result.Items = items

	return result, nil
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

	var published *time.Time
	if entry.PublishedParsed != nil {
		published = entry.PublishedParsed
	} else if entry.UpdatedParsed != nil {
		published = entry.UpdatedParsed
	}

	return &FeedItem{
		Title:     title,
		Link:      link,
		GUID:      guid,
		Published: published,
		Summary:   getSummary(entry),
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
		if items[i].Published != nil {
			if mostRecent == nil || items[i].Published.After(*mostRecent.Published) {
				mostRecent = &items[i]
			}
		}
	}

	if mostRecent != nil {
		return mostRecent
	}

	return &items[0]
}
