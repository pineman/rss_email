# Advanced Usage

## Custom Check Intervals

You can set different check intervals by modifying `config.yaml`:

```yaml
check_interval_minutes: 15  # Check every 15 minutes
```

Minimum recommended: 5 minutes (to avoid overwhelming feed servers)

## Running as a System Service

### Linux (systemd)

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

### macOS (launchd)

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
