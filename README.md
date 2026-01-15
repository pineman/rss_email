# RSS to Email

Monitors RSS/Atom feeds and sends new posts via email. Checks every 30 minutes.

This program is designed to be a "good netizen" and follows the [Feed Reader Behavior (FRB)](https://rachelbythebay.com/frb/) rules. It implements proper conditional requests (ETag/Last-Modified preservation), respects rate limiting (`Retry-After`), and uses exponential backoff for failing feeds to minimize server load.

## Installation

## Setup

Create `.env`:

```
GMAIL_APP_PASSWORD=your-app-password
```

Create `config.yaml`:

```
feeds:
  - "https://blog.example.com/feed.xml"
```

Build and run:

```bash
go build
./rss-email
```

For new feeds, only the most recent post is emailed to avoid inbox flooding.

## Gmail App Password

Generate at [Google Account settings](https://myaccount.google.com/) → Security → 2-Step Verification → App passwords.
