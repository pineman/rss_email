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

from database import initialize_database, is_item_sent, mark_item_sent, get_sent_count, has_feed_items
from email_sender import EmailSender, format_rss_email
from rss_checker import fetch_feed

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[
        logging.StreamHandler(sys.stdout)
    ]
)

logger = logging.getLogger(__name__)

# Global configuration
CONFIG = {}
EMAIL_SENDER = None


def load_configuration():
    """Load configuration from .env and config.yaml files."""
    global CONFIG, EMAIL_SENDER
    
    # Load environment variables
    load_dotenv()
    
    # Check required environment variables
    required_env_vars = ['GMAIL_ADDRESS', 'GMAIL_APP_PASSWORD', 'RECIPIENT_EMAIL']
    missing_vars = [var for var in required_env_vars if not os.getenv(var)]
    
    if missing_vars:
        logger.error(f"Missing required environment variables: {', '.join(missing_vars)}")
        logger.error("Please create a .env file based on .env.example")
        sys.exit(1)
    
    # Initialize email sender
    EMAIL_SENDER = EmailSender(
        gmail_address=os.getenv('GMAIL_ADDRESS'),
        gmail_app_password=os.getenv('GMAIL_APP_PASSWORD'),
        recipient_email=os.getenv('RECIPIENT_EMAIL')
    )
    
    # Load YAML configuration
    config_path = Path(__file__).parent / "config.yaml"
    
    if not config_path.exists():
        logger.error("config.yaml not found. Please create it based on config.example.yaml")
        sys.exit(1)
    
    try:
        with open(config_path, 'r') as f:
            CONFIG = yaml.safe_load(f)
        
        # Validate configuration
        if 'feeds' not in CONFIG or not CONFIG['feeds']:
            logger.error("No feeds configured in config.yaml")
            sys.exit(1)
        
        # Set default check interval if not specified
        if 'check_interval_minutes' not in CONFIG:
            CONFIG['check_interval_minutes'] = 30
        
        logger.info(f"Loaded configuration with {len(CONFIG['feeds'])} feed(s)")
        logger.info(f"Check interval: {CONFIG['check_interval_minutes']} minutes")
        
    except Exception as e:
        logger.error(f"Error loading config.yaml: {e}")
        sys.exit(1)


def check_feeds():
    """Check all configured RSS feeds for new items and send emails."""
    logger.info("=== Starting feed check cycle ===")
    
    feeds_checked = 0
    items_found = 0
    emails_sent = 0
    errors = 0
    
    for feed_url in CONFIG['feeds']:
        if not feed_url:
            logger.warning("Skipping empty feed URL")
            continue
        
        feed_name = feed_url  # Default to URL in case of error
        
        try:
            logger.info(f"Checking feed: {feed_url}")
            feeds_checked += 1
            
            # Fetch feed items and feed title
            feed_name, items = fetch_feed(feed_url)
            logger.info(f"Feed title: {feed_name}")
            items_found += len(items)
            
            # Check if this is a new feed (no items in database)
            is_new_feed = not has_feed_items(feed_url)
            
            if is_new_feed and items:
                logger.info(f"New feed detected: {feed_name}. Sending only the most recent post.")
                
                # Sort items by publication date to find the most recent
                # Items with dates come first, sorted by date (newest first)
                # Items without dates come last
                items_with_dates = [item for item in items if item.get('published_dt')]
                items_without_dates = [item for item in items if not item.get('published_dt')]
                
                if items_with_dates:
                    items_with_dates.sort(key=lambda x: x['published_dt'], reverse=True)
                    most_recent_item = items_with_dates[0]
                    logger.info(f"Most recent post published: {most_recent_item['published']}")
                else:
                    # No dates available, use first item from feed (feed's default order)
                    most_recent_item = items[0]
                    logger.info("No publication dates available, using first item from feed")
                try:
                    subject, text_body, html_body = format_rss_email(feed_name, most_recent_item)
                    
                    if EMAIL_SENDER.send_email(subject, text_body, html_body):
                        mark_item_sent(feed_url, most_recent_item['guid'])
                        emails_sent += 1
                        logger.info(f"Sent most recent post: {most_recent_item['title']}")
                    else:
                        logger.error(f"Failed to send email for: {most_recent_item['title']}")
                        errors += 1
                except Exception as e:
                    logger.error(f"Error processing most recent item '{most_recent_item.get('title', 'Unknown')}': {e}")
                    errors += 1
                
                # Mark all other items as sent without sending emails
                for item in items[1:]:
                    try:
                        mark_item_sent(feed_url, item['guid'])
                        logger.debug(f"Marked as sent (no email): {item['title']}")
                    except Exception as e:
                        logger.warning(f"Error marking item as sent: {e}")
                
                logger.info(f"Marked {len(items) - 1} older posts as sent without sending emails")
            else:
                # Process each item normally for existing feeds
                for item in items:
                    try:
                        # Check if item was already sent
                        if is_item_sent(feed_url, item['guid']):
                            logger.debug(f"Already sent: {item['title']}")
                            continue
                        
                        # Format and send email
                        subject, text_body, html_body = format_rss_email(feed_name, item)
                        
                        if EMAIL_SENDER.send_email(subject, text_body, html_body):
                            # Mark item as sent
                            mark_item_sent(feed_url, item['guid'])
                            emails_sent += 1
                            logger.info(f"Sent: {item['title']}")
                        else:
                            logger.error(f"Failed to send email for: {item['title']}")
                            errors += 1
                        
                        # Small delay to avoid overwhelming SMTP server
                        time.sleep(1)
                        
                    except Exception as e:
                        logger.error(f"Error processing item '{item.get('title', 'Unknown')}': {e}")
                        errors += 1
        
        except Exception as e:
            logger.error(f"Error checking feed {feed_name}: {e}")
            errors += 1
    
    # Log summary
    total_sent = get_sent_count()
    logger.info(f"=== Feed check complete ===")
    logger.info(f"Feeds checked: {feeds_checked}")
    logger.info(f"Items found: {items_found}")
    logger.info(f"Emails sent this cycle: {emails_sent}")
    logger.info(f"Errors: {errors}")
    logger.info(f"Total items sent (all time): {total_sent}")


def log_heartbeat():
    """Log a periodic heartbeat message."""
    logger.info("❤️ Application is running")


def main():
    """Main application entry point."""
    logger.info("=== RSS to Email Service Starting ===")
    
    # Load configuration
    load_configuration()
    
    # Initialize database
    initialize_database()
    
    # Do an initial check immediately
    logger.info("Performing initial feed check...")
    try:
        check_feeds()
    except Exception as e:
        logger.error(f"Error during initial feed check: {e}")
    
    # Set up scheduler
    scheduler = BlockingScheduler()
    
    # Schedule feed checks
    check_interval = CONFIG['check_interval_minutes']
    scheduler.add_job(
        check_feeds,
        trigger=IntervalTrigger(minutes=check_interval),
        id='check_feeds',
        name='Check RSS feeds',
        replace_existing=True
    )
    
    # Schedule hourly heartbeat
    scheduler.add_job(
        log_heartbeat,
        trigger=IntervalTrigger(hours=1),
        id='heartbeat',
        name='Heartbeat log',
        replace_existing=True
    )
    
    logger.info(f"Scheduler started. Checking feeds every {check_interval} minutes.")
    logger.info("Press Ctrl+C to stop.")
    
    try:
        scheduler.start()
    except KeyboardInterrupt:
        logger.info("Shutting down gracefully...")
        scheduler.shutdown()
        logger.info("=== RSS to Email Service Stopped ===")


if __name__ == "__main__":
    main()
