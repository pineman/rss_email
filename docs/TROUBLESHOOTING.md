# Troubleshooting

## SMTP Authentication Failed

**Problem**: "SMTP authentication failed" error

**Solutions**:
- Verify your Gmail address is correct
- Ensure you're using an App Password, not your regular Gmail password
- Check that 2-Step Verification is enabled on your Google Account
- Regenerate the App Password if needed

## Feed Not Updating

**Problem**: No new items being sent despite new posts in feed

**Solutions**:
- Check the feed URL is correct and accessible
- Verify the `check_interval_minutes` is appropriate
- Check logs for feed parsing errors
- Some feeds may have delays before publishing new items

## Adding a New Feed

**Behavior**: When you add a new feed to `config.yaml`

The application will:
- Detect that it's a new feed (no items in database)
- Sort all posts by publication date to find the most recent one
- Send an email for only the most recent post
- Mark all older posts as "sent" without sending emails
- Going forward, only send emails for new posts

This prevents your inbox from being flooded with old posts when adding a feed.

## Database Locked Error

**Problem**: "Database is locked" error

**Solutions**:
- Ensure only one instance of the application is running
- Check file permissions on the `data/` directory
- If using Docker, ensure the volume is properly mounted

## No Emails Received

**Problem**: Application runs but no emails arrive

**Solutions**:
- Check spam/junk folder
- Verify `RECIPIENT_EMAIL` is correct in `.env`
- Check logs for email sending errors
- Ensure your Gmail account can send emails
- Test with a simple feed that has recent posts
