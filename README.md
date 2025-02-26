# Google Kit - Model Context Protocol (MCP) Server

The Model Context Protocol (MCP) implementation in Google Kit enables AI models to interact with Google services through a standardized interface.

## Prerequisites

- Go 1.23.2 or higher
- Google Cloud Platform account with the following APIs enabled:
  - Google Calendar API
  - Gmail API
  - Google Chat API
- Google Cloud credentials (OAuth 2.0 Client ID)

## Installation

### Installing via Go

1. Install the server:

```bash
go install github.com/nguyenvanduocit/google-kit@latest
```

2. Create a `.env` file with your configuration:

```env
# Required for Google Services
GOOGLE_CREDENTIALS_FILE=  # Required: Path to Google Cloud credentials JSON file
GOOGLE_TOKEN_FILE=       # Required: Path to store Google OAuth tokens

# Optional configurations
ENABLE_TOOLS=           # Optional: Comma-separated list of tool groups to enable (empty = all enabled)
PROXY_URL=             # Optional: HTTP/HTTPS proxy URL if needed
```

https://developers.google.com/workspace/chat/authenticate-authorize-chat-user

3. Configure your Claude's config:

```json
{
  "mcpServers": {
    "google_kit": {
      "command": "google-kit",
      "args": ["-env", "/path/to/.env"]
    }
  }
}
```

## Enable Tools

The `ENABLE_TOOLS` environment variable is a comma-separated list of tool groups to enable. Available groups are:
- `calendar` - Google Calendar tools
- `gmail` - Gmail tools
- `gchat` - Google Chat tools

Leave it empty to enable all tools.

## Available Tools

### Group: calendar

#### calendar_create_event
Create a new event in Google Calendar with title, description, time, and attendees.

#### calendar_list_events
List upcoming events in Google Calendar with customizable time range and result limit.

#### calendar_update_event
Update an existing event's details including title, description, time, and attendees.

#### calendar_respond_to_event
Respond to an event invitation (accept, decline, or tentative).

### Group: gchat

#### gchat_list_spaces
List all available Google Chat spaces/rooms.

#### gchat_send_message
Send a message to a Google Chat space or direct message.

### Group: gmail

#### gmail_search
Search emails in Gmail using Gmail's search syntax.

#### gmail_move_to_spam
Move specific emails to spam folder in Gmail by message IDs.

#### gmail_create_filter
Create a Gmail filter with specified criteria and actions:
- Filter by sender, recipient, subject, or custom query
- Add labels
- Mark as important
- Mark as read
- Archive messages

#### gmail_list_filters
List all Gmail filters in the account.

#### gmail_list_labels
List all Gmail labels in the account.

#### gmail_delete_filter
Delete a Gmail filter by its ID.

#### gmail_delete_label
Delete a Gmail label by its ID.

