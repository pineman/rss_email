"""Database operations module for tracking sent RSS items."""

import sqlite3
import os
import logging
from datetime import datetime
from pathlib import Path

logger = logging.getLogger(__name__)

DB_DIR = Path(__file__).parent / "data"
DB_PATH = DB_DIR / "rss_email.db"


def initialize_database():
    DB_DIR.mkdir(exist_ok=True)
    
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS sent_items (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            feed_url TEXT NOT NULL,
            item_guid TEXT NOT NULL,
            sent_at TIMESTAMP NOT NULL,
            UNIQUE(feed_url, item_guid)
        )
    """)
    
    cursor.execute("""
        CREATE INDEX IF NOT EXISTS idx_feed_guid 
        ON sent_items(feed_url, item_guid)
    """)
    
    conn.commit()
    conn.close()
    
    logger.info(f"Database initialized at {DB_PATH}")


def is_item_sent(feed_url: str, item_guid: str) -> bool:
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
        logger.warning(f"Item already marked as sent: {item_guid} from {feed_url}")
    finally:
        conn.close()


def get_sent_count() -> int:
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute("SELECT COUNT(*) FROM sent_items")
    count = cursor.fetchone()[0]
    
    conn.close()
    
    return count


def has_feed_items(feed_url: str) -> bool:
    conn = sqlite3.connect(DB_PATH)
    cursor = conn.cursor()
    
    cursor.execute(
        "SELECT COUNT(*) FROM sent_items WHERE feed_url = ?",
        (feed_url,)
    )
    
    count = cursor.fetchone()[0]
    conn.close()
    
    return count > 0


