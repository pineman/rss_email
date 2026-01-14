"""RSS operations module for fetching and parsing RSS/Atom feeds."""

import feedparser
import logging
import ssl
from typing import List, Dict, Optional, Tuple
from datetime import datetime

# Disable SSL certificate verification for feedparser
# This is needed for feeds with SSL certificate issues
if hasattr(ssl, '_create_unverified_context'):
    ssl._create_default_https_context = ssl._create_unverified_context

logger = logging.getLogger(__name__)


def fetch_feed(url: str) -> Tuple[str, List[Dict]]:
    """
    Fetch and parse an RSS/Atom feed.
    
    Args:
        url: The URL of the RSS feed
        
    Returns:
        Tuple of (feed_title, items) where:
        - feed_title: The title of the feed itself
        - items: List of normalized item dictionaries with keys:
            - title: Item title
            - link: Item link/URL
            - guid: Unique identifier (guid or link as fallback)
            - published: Publication date (formatted string)
            - summary: Item description/summary
    """
    items = []
    feed_title = url  # Default to URL if no title found
    
    logger.debug(f"Fetching feed: {url}")
    feed = feedparser.parse(url)
    
    if feed.bozo and hasattr(feed, 'bozo_exception'):
        logger.warning(f"Feed parsing warning for {url}: {feed.bozo_exception}")
    
    if hasattr(feed, 'status') and feed.status >= 400:
        logger.error(f"HTTP error {feed.status} when fetching {url}")
        return feed_title, items
    
    if hasattr(feed, 'feed') and hasattr(feed.feed, 'title'):
        feed_title = feed.feed.title
    
    for entry in feed.entries:
        item = normalize_feed_item(entry)
        if item:
            items.append(item)
    
    logger.info(f"Fetched {len(items)} items from {feed_title} ({url})")
    
    return feed_title, items


def normalize_feed_item(entry) -> Optional[Dict]:
    title = entry.get('title', 'No Title')
    link = entry.get('link', '')
    guid = entry.get('id', entry.get('guid', link))  # id -> guid -> link fallback
    
    if not guid:
        logger.warning("Skipping item with no guid or link")
        return None
    
    published = get_published_date(entry)
    published_dt = get_published_datetime(entry)
    summary = get_summary(entry)
    
    return {
        'title': title,
        'link': link,
        'guid': guid,
        'published': published,
        'published_dt': published_dt,  # For sorting
        'summary': summary
    }


def get_published_date(entry) -> str:
    date_fields = ['published_parsed', 'updated_parsed', 'created_parsed']
    
    for field in date_fields:
        if hasattr(entry, field):
            date_tuple = getattr(entry, field)
            if date_tuple:
                try:
                    dt = datetime(*date_tuple[:6])
                    return dt.strftime('%Y-%m-%d %H:%M:%S')
                except Exception:
                    pass
    
    for field in ['published', 'updated', 'created']:
        if hasattr(entry, field):
            date_str = getattr(entry, field)
            if date_str:
                return date_str
    
    return 'Unknown'


def get_published_datetime(entry) -> Optional[datetime]:
    date_fields = ['published_parsed', 'updated_parsed', 'created_parsed']
    
    for field in date_fields:
        if hasattr(entry, field):
            date_tuple = getattr(entry, field)
            if date_tuple:
                try:
                    return datetime(*date_tuple[:6])
                except Exception:
                    pass
    
    return None


def get_summary(entry) -> str:
    if hasattr(entry, 'summary') and entry.summary:
        return entry.summary
    
    if hasattr(entry, 'description') and entry.description:
        return entry.description
    
    if hasattr(entry, 'content') and entry.content:
        if isinstance(entry.content, list) and len(entry.content) > 0:
            return entry.content[0].get('value', 'No summary available.')
    
    return 'No summary available.'


def validate_feed_url(url: str) -> bool:
    feed = feedparser.parse(url)
    
    if hasattr(feed, 'entries') and len(feed.entries) >= 0:
        return True
    
    return False
