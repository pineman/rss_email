# RSS to Email Service

A Python application that monitors RSS/Atom feeds and sends new posts via email to a Gmail address. Perfect for staying updated with your favorite blogs, news sites, and content creators without constantly checking multiple feeds.

## Features

- 📰 Monitors multiple RSS/Atom feeds
- 📧 Sends beautifully formatted HTML emails via Gmail
- 🗄️ Tracks sent items to avoid duplicates using SQLite
- ⚙️ Configurable check intervals
- 🐳 Docker support for easy deployment
- 🔄 Automatic retry and error handling
- 📊 Comprehensive logging
- 🆕 Smart new feed handling - only sends the most recent post when adding a new feed

## Prerequisites

- Python 3.11 or higher (for local installation)
- Gmail account with App Password enabled
- Docker (optional, for containerized deployment)

## Quick Start

### 1. Generate Gmail App Password

To send emails via Gmail SMTP, you need an App Password:

1. Go to your [Google Account settings](https://myaccount.google.com/)
2. Navigate to **Security** → **2-Step Verification** (enable if not already enabled)
3. Scroll down to **App passwords**
4. Select **Mail** as the app and **Other** as the device
5. Enter a name like "RSS to Email" and click **Generate**
6. Copy the 16-character password (you'll need it for the `.env` file)

### 2. Local Setup

```bash
# Clone or download the project
cd rss_email

# Create and activate virtual environment
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Create configuration files
cp .env.example .env
cp config.example.yaml config.yaml

# Edit .env with your email credentials
# Edit config.yaml with your RSS feeds
```

### 3. Configure Environment Variables

Edit `.env` file:

```env
GMAIL_ADDRESS=your-email@gmail.com
GMAIL_APP_PASSWORD=your-16-char-app-password
RECIPIENT_EMAIL=recipient@gmail.com
```

### 4. Configure RSS Feeds

Edit `config.yaml`:

```yaml
feeds:
  - "https://blog.example.com/feed.xml"
  - "https://news.example.com/rss"

check_interval_minutes: 30
```

**Notes:**
- The feed name is automatically extracted from each feed's title, so you only need to provide the URL.
- When you add a new feed, only the most recent post will be emailed (determined by publication date). All older posts are marked as seen to avoid flooding your inbox.

### 5. Run the Application

```bash
python app.py
```

The application will:
1. Perform an initial check of all feeds
2. Send emails for any new items found
3. Continue running and check feeds at the configured interval
4. Press `Ctrl+C` to stop

## Docker Deployment

### Using Docker Compose (Recommended)

```bash
# Create .env and config.yaml as described above

# Build and run
docker-compose up -d

# View logs
docker-compose logs -f

# Stop the service
docker-compose down
```

### Using Docker CLI

```bash
# Build the image
docker build -t rss-email .

# Run the container
docker run -d \
  --name rss-email-service \
  --env-file .env \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  rss-email

# View logs
docker logs -f rss-email-service

# Stop the container
docker stop rss-email-service
docker rm rss-email-service
```

## Configuration

### config.yaml

- `feeds`: List of RSS feed URLs to monitor (feed names are automatically extracted from feed titles)
- `check_interval_minutes`: How often to check feeds (default: 30 minutes)

### Environment Variables (.env)

- `GMAIL_ADDRESS`: Your Gmail address (sender)
- `GMAIL_APP_PASSWORD`: Gmail app password (NOT your regular password)
- `RECIPIENT_EMAIL`: Email address to receive RSS updates

## Project Structure

```
rss_email/
├── app.py              # Main application
├── database.py         # SQLite database operations
├── email_sender.py     # Email sending functionality
├── rss_checker.py      # RSS feed parsing
├── requirements.txt    # Python dependencies
├── config.yaml         # RSS feeds configuration
├── .env               # Environment variables (not in git)
├── data/              # SQLite database storage (not in git)
├── Dockerfile         # Docker configuration
├── docker-compose.yml # Docker Compose configuration
└── README.md          # This file
```

## How It Works

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

## Troubleshooting

### SMTP Authentication Failed

**Problem**: "SMTP authentication failed" error

**Solutions**:
- Verify your Gmail address is correct
- Ensure you're using an App Password, not your regular Gmail password
- Check that 2-Step Verification is enabled on your Google Account
- Regenerate the App Password if needed

### Feed Not Updating

**Problem**: No new items being sent despite new posts in feed

**Solutions**:
- Check the feed URL is correct and accessible
- Verify the `check_interval_minutes` is appropriate
- Check logs for feed parsing errors
- Some feeds may have delays before publishing new items

### Adding a New Feed

**Behavior**: When you add a new feed to `config.yaml`

The application will:
- Detect that it's a new feed (no items in database)
- Sort all posts by publication date to find the most recent one
- Send an email for only the most recent post
- Mark all older posts as "sent" without sending emails
- Going forward, only send emails for new posts

This prevents your inbox from being flooded with old posts when adding a feed.

### Database Locked Error

**Problem**: "Database is locked" error

**Solutions**:
- Ensure only one instance of the application is running
- Check file permissions on the `data/` directory
- If using Docker, ensure the volume is properly mounted

### No Emails Received

**Problem**: Application runs but no emails arrive

**Solutions**:
- Check spam/junk folder
- Verify `RECIPIENT_EMAIL` is correct in `.env`
- Check logs for email sending errors
- Ensure your Gmail account can send emails
- Test with a simple feed that has recent posts

## Advanced Usage

### Custom Check Intervals

You can set different check intervals by modifying `config.yaml`:

```yaml
check_interval_minutes: 15  # Check every 15 minutes
```

Minimum recommended: 5 minutes (to avoid overwhelming feed servers)

### Running as a System Service

#### Linux (systemd)

Create `/etc/systemd/system/rss-email.service`:

```ini
[Unit]
Description=RSS to Email Service
After=network.target

[Service]
Type=simple
User=your-username
WorkingDirectory=/path/to/rss_email
ExecStart=/path/to/venv/bin/python app.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl enable rss-email
sudo systemctl start rss-email
sudo systemctl status rss-email
```

#### macOS (launchd)

Create `~/Library/LaunchAgents/com.user.rss-email.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.rss-email</string>
    <key>ProgramArguments</key>
    <array>
        <string>/path/to/venv/bin/python</string>
        <string>/path/to/rss_email/app.py</string>
    </array>
    <key>WorkingDirectory</key>
    <string>/path/to/rss_email</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
</dict>
</plist>
```

Then:
```bash
launchctl load ~/Library/LaunchAgents/com.user.rss-email.plist
```

## Maintenance

### Database Cleanup

The database grows over time but remains small. To clean up old entries:

```python
from database import cleanup_old_entries

# Remove entries older than 90 days
cleanup_old_entries(days=90)
```

You can add this to a scheduled task or run it manually when needed.

### Monitoring

Check the application is running:
- Look for hourly heartbeat messages in logs
- Monitor the `data/rss_email.db` file size
- Check email delivery

## Security Notes

- Never commit `.env` file to version control
- Keep your App Password secure
- Use App Passwords instead of regular Gmail password
- Restrict file permissions on `.env`: `chmod 600 .env`
- Regularly update Python dependencies

## Dependencies

- `feedparser==6.0.11` - RSS/Atom feed parsing
- `python-dotenv==1.0.0` - Environment variable management
- `APScheduler==3.10.4` - Periodic task scheduling
- `PyYAML==6.0.1` - YAML configuration parsing

## License

This project is provided as-is for personal use.

## Support

For issues, questions, or contributions:
1. Check the troubleshooting section above
2. Review application logs for error messages
3. Verify configuration files are correct
4. Ensure all dependencies are installed

## Future Enhancements

Potential features for future versions:
- Web interface for feed management
- Multiple recipient support
- Digest mode (batch multiple items into one email)
- Custom email templates
- Feed-specific filters and rules
- Webhook notifications
- RSS feed validation and testing tool
