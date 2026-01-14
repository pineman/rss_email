"""Email operations module for sending RSS items via Gmail SMTP."""

import smtplib
import logging
from email.mime.text import MIMEText
from email.mime.multipart import MIMEMultipart
from typing import Optional

logger = logging.getLogger(__name__)

# Gmail SMTP configuration
SMTP_SERVER = "smtp.gmail.com"
SMTP_PORT = 587


class EmailSender:
    """Email sender class for Gmail SMTP."""
    
    def __init__(self, gmail_address: str, gmail_app_password: str, recipient_email: str):
        """
        Initialize the email sender.
        
        Args:
            gmail_address: Sender's Gmail address
            gmail_app_password: Gmail app password (not regular password)
            recipient_email: Recipient's email address
        """
        self.gmail_address = gmail_address
        self.gmail_app_password = gmail_app_password
        self.recipient_email = recipient_email
        
    def send_email(self, subject: str, body: str, html_body: Optional[str] = None) -> bool:
        """
        Send an email via Gmail SMTP.
        
        Args:
            subject: Email subject line
            body: Plain text email body
            html_body: Optional HTML version of the email body
            
        Returns:
            True if email was sent successfully, False otherwise
        """
        try:
            # Create message
            if html_body:
                msg = MIMEMultipart("alternative")
                msg.attach(MIMEText(body, "plain"))
                msg.attach(MIMEText(html_body, "html"))
            else:
                msg = MIMEText(body, "plain")
            
            msg["Subject"] = subject
            msg["From"] = self.gmail_address
            msg["To"] = self.recipient_email
            
            # Connect to Gmail SMTP server
            with smtplib.SMTP(SMTP_SERVER, SMTP_PORT) as server:
                server.starttls()  # Upgrade connection to TLS
                server.login(self.gmail_address, self.gmail_app_password)
                server.send_message(msg)
            
            logger.info(f"Email sent successfully: {subject}")
            return True
            
        except smtplib.SMTPAuthenticationError:
            logger.error("SMTP authentication failed. Check your Gmail address and app password.")
            return False
        except smtplib.SMTPException as e:
            logger.error(f"SMTP error occurred: {e}")
            return False
        except Exception as e:
            logger.error(f"Failed to send email: {e}")
            return False


def format_rss_email(feed_name: str, item: dict) -> tuple[str, str, str]:
    """
    Format an RSS item into email subject and body.
    
    Args:
        feed_name: Name of the RSS feed
        item: Dictionary containing RSS item data (title, link, published, summary)
        
    Returns:
        Tuple of (subject, text_body, html_body)
    """
    # Email subject
    subject = f"[RSS] {feed_name}: {item.get('title', 'No Title')}"
    
    # Plain text body
    text_body = f"""
New post from {feed_name}

Title: {item.get('title', 'No Title')}
Link: {item.get('link', 'No Link')}
Published: {item.get('published', 'Unknown')}

{item.get('summary', 'No summary available.')}

---
This email was sent by RSS to Email service.
"""
    
    # HTML body
    html_body = f"""
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2 style="color: #2c3e50;">New post from {feed_name}</h2>
    
    <div style="background-color: #f8f9fa; padding: 15px; border-left: 4px solid #3498db; margin: 20px 0;">
        <h3 style="margin-top: 0;">
            <a href="{item.get('link', '#')}" style="color: #2980b9; text-decoration: none;">
                {item.get('title', 'No Title')}
            </a>
        </h3>
        <p style="color: #7f8c8d; font-size: 0.9em;">
            <strong>Published:</strong> {item.get('published', 'Unknown')}
        </p>
    </div>
    
    <div style="margin: 20px 0;">
        {item.get('summary', 'No summary available.')}
    </div>
    
    <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #ecf0f1;">
        <p style="color: #95a5a6; font-size: 0.85em;">
            This email was sent by RSS to Email service.
        </p>
    </div>
</body>
</html>
"""
    
    return subject, text_body, html_body
