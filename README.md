# RSS to Email

Monitors RSS/Atom feeds and sends new posts via email. Checks every 30 minutes.

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
