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
├── docs/              # Additional documentation
├── Dockerfile         # Docker configuration
├── docker-compose.yml # Docker Compose configuration
└── README.md          # This file
```

## Documentation

For more detailed information, see:
- [How It Works](docs/HOW_IT_WORKS.md) - Technical details on how the application works
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues and solutions
- [Advanced Usage](docs/ADVANCED_USAGE.md) - System service setup, maintenance, and security

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
