"""Database operations module for tracking sent RSS items."""

import sqlite3
import os
import logging
from datetime import datetime
from pathlib import Path

logger = logging.getLogger(__name__)

# Database path
DB_DIR = Path(__file__).parent / "data"
DB_PATH = DB_DIR / "rss_email.db"


def initialize_database():
    """Initialize the SQLite database and create tables if they don't exist."""
    # Ensure data directory exists
    DB_DIR.mkdir(exist_ok=True)
    
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    # Create sent_items table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS sent_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            feed_url TEXT NOT NULL,
            item_guid TEXT NOT NULL,
            sent_at TIMESTAMP NOT NULL,
            UNIQUE(feed_url, item_guid)
        )
    """)
    
    # Create index for faster lookups
    cursor.execute("""
        CREATE INDEX IF NOT EXISTS idx_feed_guid 
        ON sent_items(feed_url, item_guid)
    """)
    
    conn.commit()
    conn.close()
    
    logger.info(f"Database initialized at {DB_PATH}")


def is_item_sent(feed_url: str, item_guid: str) -> bool:
    """
    Check if an RSS item has already been sent.
    
    Args:
        feed_url: The URL of the RSS feed
        item_guid: The unique identifier (guid or link) of the item
        
    Returns:
        True if the item has been sent, False otherwise
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute(
        "SELECT COUNT(*) FROM sent_items WHERE feed_url = ? AND item_guid = ?",
        (feed_url, item_guid)
    )
    
    count = cursor.fetchone()[0]
    conn.close()
    
    return count > 0


def mark_item_sent(feed_url: str, item_guid: str):
    """
    Mark an RSS item as sent by adding it to the database.
    
    Args:
        feed_url: The URL of the RSS feed
        item_guid: The unique identifier (guid or link) of the item
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    try:
        cursor.execute(
            "INSERT INTO sent_items (feed_url, item_guid, sent_at) VALUES (?, ?, ?)",
            (feed_url, item_guid, datetime.now())
        )
        conn.commit()
        logger.debug(f"Marked item as sent: {item_guid} from {feed_url}")
    except sqlite3.IntegrityError:
        # Item already exists in database
        logger.warning(f"Item already marked as sent: {item_guid} from {feed_url}")
    finally:
        conn.close()


def get_sent_count() -> int:
    """
    Get the total number of sent items.
    
    Returns:
        The count of items in the database
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute("SELECT COUNT(*) FROM sent_items")
    count = cursor.fetchone()[0]
    
    conn.close()
    
    return count


def has_feed_items(feed_url: str) -> bool:
    """
    Check if any items from a feed have been recorded in the database.
    
    Args:
        feed_url: The URL of the RSS feed
        
    Returns:
        True if the feed has any items in the database, False otherwise
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute(
        "SELECT COUNT(*) FROM sent_items WHERE feed_url = ?",
        (feed_url,)
    )
    
    count = cursor.fetchone()[0]
    conn.close()
    
    return count > 0


def cleanup_old_entries(days: int = 90):
    """
    Remove entries older than the specified number of days.
    
    Args:
        days: Number of days to keep (default: 90)
    """
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute(
        "DELETE FROM sent_items WHERE sent_at < datetime('now', ? || ' days')",
        (f"-{days}",)
    )
    
    deleted_count = cursor.rowcount
    conn.commit()
    conn.close()
    
    logger.info(f"Cleaned up {deleted_count} old entries (older than {days} days)")
    
    return deleted_count
