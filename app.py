"""Main application for RSS to Email service."""

import os
import sys
import logging
import yaml
import time
from pathlib import Path
from dotenv import load_dotenv
from apscheduler.schedulers.blocking import BlockingScheduler
from apscheduler.triggers.interval import IntervalTrigger

from database import initialize_database, is_item_sent, mark_item_sent, has_feed_items
from email_sender import EmailSender, format_rss_email
from rss_checker import fetch_feed

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s",
    handlers=[logging.StreamHandler(sys.stdout)],
)

logger = logging.getLogger(__name__)

CONFIG = {}
EMAIL_SENDER = None


def load_configuration():
    """Load configuration from .env and config.yaml files."""
    global CONFIG, EMAIL_SENDER

    load_dotenv()

    required_env_vars = ["GMAIL_ADDRESS", "GMAIL_APP_PASSWORD", "RECIPIENT_EMAIL"]
    missing_vars = [var for var in required_env_vars if not os.getenv(var)]

    if missing_vars:
        logger.error(f"Missing required environment variables: {', '.join(missing_vars)}")
        logger.error("Please create a .env file based on .env.example")
        sys.exit(1)

    EMAIL_SENDER = EmailSender(
        gmail_address=os.getenv("GMAIL_ADDRESS"),
        gmail_app_password=os.getenv("GMAIL_APP_PASSWORD"),
        recipient_email=os.getenv("RECIPIENT_EMAIL"),
    )

    config_path = Path(__file__).parent / "config.yaml"

    if not config_path.exists():
        logger.error("config.yaml not found. Please create it based on config.example.yaml")
        sys.exit(1)

    with open(config_path, "r") as f:
        CONFIG = yaml.safe_load(f)

    if "feeds" not in CONFIG or not CONFIG["feeds"]:
        logger.error("No feeds configured in config.yaml")
        sys.exit(1)


def send_item(feed_url, feed_name, item):
    """Send an email for an item and mark it as sent."""
    subject, text_body, html_body = format_rss_email(feed_name, item)
    EMAIL_SENDER.send_email(subject, text_body, html_body)
    mark_item_sent(feed_url, item["guid"])
    logger.info(f"Sent: {item['title']} for {feed_url}")
    time.sleep(1)


def get_most_recent_item(items):
    """Get the most recent item by publication date, or first item if no dates."""
    items_with_dates = [item for item in items if item.get("published_dt")]
    if items_with_dates:
        return max(items_with_dates, key=lambda x: x["published_dt"])
    return items[0]


def process_new_feed(feed_url, feed_name, items):
    """For new feeds, send only the most recent post and mark all others as sent."""
    most_recent = get_most_recent_item(items)
    send_item(feed_url, feed_name, most_recent)
    for item in items:
        if item["guid"] != most_recent["guid"]:
            mark_item_sent(feed_url, item["guid"])


def process_existing_feed(feed_url, feed_name, items):
    """For existing feeds, send all unsent items."""
    for item in items:
        if not is_item_sent(feed_url, item["guid"]):
            send_item(feed_url, feed_name, item)


def check_feeds():
    """Check all configured RSS feeds for new items and send emails."""
    logger.info("Checking feeds...")
    for feed_url in CONFIG["feeds"]:
        feed_name, items = fetch_feed(feed_url)

        if not items:
            continue

        if has_feed_items(feed_url):
            process_existing_feed(feed_url, feed_name, items)
        else:
            process_new_feed(feed_url, feed_name, items)
    logger.info("Done checking feeds.")

def main():
    load_configuration()
    initialize_database()
    check_feeds()

    scheduler = BlockingScheduler()
    scheduler.add_job(
        check_feeds,
        trigger=IntervalTrigger(minutes=30),
        id="check_feeds",
        name="Check RSS feeds",
        replace_existing=True,
    )

    logger.info("Scheduler started")

    try:
        scheduler.start()
    except KeyboardInterrupt:
        logger.info("Shutting down")
        scheduler.shutdown()


if __name__ == "__main__":
    main()
