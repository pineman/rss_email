# How It Works

1. **Initialization**: The application loads configuration and initializes the SQLite database
2. **Feed Checking**: At configured intervals, it fetches each RSS feed
3. **New Feed Detection**: When a feed is added for the first time:
   - Only the most recent post is sent via email
   - All older posts are marked as sent (without sending emails) to avoid spam
4. **Duplicate Detection**: Each item's GUID is checked against the database
5. **Email Sending**: New items are formatted and sent via Gmail SMTP
6. **Tracking**: Sent items are recorded in the database to prevent duplicates
7. **Logging**: All activities are logged to the console

## Database

The application uses SQLite to track sent items:

- Database location: `./data/rss_email.db`
- Table: `sent_items(feed_url, item_guid, sent_at)`
- Automatically created on first run
- Persists across restarts to prevent duplicate emails

## Logging

The application logs to stdout with the following information:

- Feed check cycles
- Items found and emails sent
- Errors and warnings
- Hourly heartbeat messages
- Statistics (feeds checked, items found, emails sent)

Log levels:
- `INFO`: Normal operations
- `WARNING`: Non-critical issues
- `ERROR`: Failures that need attention
- `DEBUG`: Detailed debugging information (disabled by default)
