"""RSS operations module for fetching and parsing RSS/Atom feeds."""

import feedparser
import logging
from typing import List, Dict, Optional, Tuple
from datetime import datetime

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
    
    try:
        logger.debug(f"Fetching feed: {url}")
        feed = feedparser.parse(url)
        
        # Check if feed was fetched successfully
        if feed.bozo and hasattr(feed, 'bozo_exception'):
            logger.warning(f"Feed parsing warning for {url}: {feed.bozo_exception}")
        
        # Check for HTTP errors
        if hasattr(feed, 'status'):
            if feed.status >= 400:
                logger.error(f"HTTP error {feed.status} when fetching {url}")
                return feed_title, items
        
        # Extract feed title
        if hasattr(feed, 'feed') and hasattr(feed.feed, 'title'):
            feed_title = feed.feed.title
        
        # Parse feed entries
        for entry in feed.entries:
            item = normalize_feed_item(entry)
            if item:
                items.append(item)
        
        logger.info(f"Fetched {len(items)} items from {feed_title} ({url})")
        
    except Exception as e:
        logger.error(f"Error fetching feed {url}: {e}")
    
    return feed_title, items


def normalize_feed_item(entry) -> Optional[Dict]:
    """
    Normalize a feed entry into a consistent dictionary format.
    
    Args:
        entry: feedparser entry object
        
    Returns:
        Dictionary with normalized item data, or None if essential data is missing
    """
    try:
        # Get title
        title = entry.get('title', 'No Title')
        
        # Get link
        link = entry.get('link', '')
        
        # Get guid (use id first, then guid, then link as fallback)
        guid = entry.get('id', entry.get('guid', link))
        
        # If no guid or link, skip this item
        if not guid:
            logger.warning("Skipping item with no guid or link")
            return None
        
        # Get published date
        published = get_published_date(entry)
        published_dt = get_published_datetime(entry)
        
        # Get summary/description
        summary = get_summary(entry)
        
        return {
            'title': title,
            'link': link,
            'guid': guid,
            'published': published,
            'published_dt': published_dt,  # For sorting
            'summary': summary
        }
        
    except Exception as e:
        logger.warning(f"Error normalizing feed item: {e}")
        return None


def get_published_date(entry) -> str:
    """
    Extract and format the published date from a feed entry.
    
    Args:
        entry: feedparser entry object
        
    Returns:
        Formatted date string or 'Unknown'
    """
    # Try different date fields
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
    
    # Try string date fields
    for field in ['published', 'updated', 'created']:
        if hasattr(entry, field):
            date_str = getattr(entry, field)
            if date_str:
                return date_str
    
    return 'Unknown'


def get_published_datetime(entry) -> Optional[datetime]:
    """
    Extract the published date as a datetime object for sorting.
    
    Args:
        entry: feedparser entry object
        
    Returns:
        datetime object or None if no date found
    """
    # Try different date fields
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
    """
    Extract the summary/description from a feed entry.
    
    Args:
        entry: feedparser entry object
        
    Returns:
        Summary text or 'No summary available.'
    """
    # Try different summary fields
    if hasattr(entry, 'summary') and entry.summary:
        return entry.summary
    
    if hasattr(entry, 'description') and entry.description:
        return entry.description
    
    if hasattr(entry, 'content') and entry.content:
        # content is usually a list of dictionaries
        if isinstance(entry.content, list) and len(entry.content) > 0:
            return entry.content[0].get('value', 'No summary available.')
    
    return 'No summary available.'


def validate_feed_url(url: str) -> bool:
    """
    Validate that a URL is a valid RSS/Atom feed.
    
    Args:
        url: The URL to validate
        
    Returns:
        True if the URL appears to be a valid feed, False otherwise
    """
    try:
        feed = feedparser.parse(url)
        
        # Check if feed has entries or is a valid feed format
        if hasattr(feed, 'entries') and len(feed.entries) >= 0:
            return True
        
        return False
        
    except Exception as e:
        logger.error(f"Error validating feed URL {url}: {e}")
        return False
