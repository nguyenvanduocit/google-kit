# Technology Stack

## Architecture

The Google Kit follows a modular MCP (Model Context Protocol) server architecture:
- **Server Layer**: MCP protocol handler and request routing
- **Tools Layer**: Individual Google service implementations (Calendar, Gmail, Chat)
- **Services Layer**: Google API client management and OAuth handling  
- **Utilities Layer**: Common functionality and request processing

## Language & Runtime

### Go 1.23.2
- **Primary Language**: Go for high performance, concurrency, and strong typing
- **Minimum Version**: Go 1.23.2 required
- **Build Configuration**: CGO disabled for static binaries
- **Module**: `github.com/nguyenvanduocit/google-kit`

## Core Dependencies

### MCP Protocol
- **Library**: `github.com/mark3labs/mcp-go v0.6.0`
- **Purpose**: Model Context Protocol server implementation
- **Features**: Logging, prompt capabilities, resource capabilities

### Google APIs
- **OAuth2**: `golang.org/x/oauth2 v0.24.0` - Authentication and authorization
- **Google API Client**: `google.golang.org/api v0.197.0` - Unified Google services client
- **Supported APIs**: Calendar, Gmail, Chat, YouTube (with extensive scopes)

### Configuration Management
- **Environment Variables**: `github.com/joho/godotenv v1.5.1` - .env file loading
- **YAML Support**: `gopkg.in/yaml.v3 v3.0.1` - Configuration parsing

## Development Environment

### Build Tools
- **Just**: Task runner for development commands
- **Go Modules**: Dependency management with go.mod/go.sum

### Required Tools
- **Go 1.23.2+**: Runtime and compilation
- **Just**: Build automation (optional but recommended)
- **Git**: Version control
- **TruffleHog**: Security scanning for sensitive data

## Common Commands

### Development Workflow
```bash
# Build optimized binary
just build

# Install to GOPATH
just install  

# Update documentation
just docs

# Security scan
just scan

# Manual build
go build -ldflags="-s -w" -o ./bin/dev-kit ./main.go

# Run development server
go run main.go -env /path/to/.env
```

### Testing & Quality
```bash
# Run tests (if available)
go test ./...

# Check modules
go mod tidy

# Security scanning
trufflehog git file://. --only-verified
```

## Environment Variables

### Required Configuration
- **GOOGLE_CREDENTIALS_FILE**: Path to Google OAuth2 credentials JSON
- **GOOGLE_TOKEN_FILE**: Path to store/read Google OAuth tokens

### Optional Configuration  
- **ENABLE_TOOLS**: Comma-separated list of tool groups (`calendar`, `gmail`, `gchat`)
- **PROXY_URL**: HTTP/HTTPS proxy URL for network requests

### Example .env
```env
GOOGLE_CREDENTIALS_FILE=./bin/google-credentials.json
GOOGLE_TOKEN_FILE=./bin/google-token.json
ENABLE_TOOLS=calendar,gmail,gchat
PROXY_URL=http://proxy.example.com:8080
```

## Port Configuration

### MCP Server
- **Protocol**: Stdio-based communication (no network ports)
- **Transport**: Standard input/output streams  
- **Integration**: Connected via MCP client configuration

### Development Setup
No network ports required - the server communicates through stdio with the MCP client.

## Authentication & Security

### OAuth 2.0 Flow
- **Type**: Desktop application OAuth flow
- **Scopes**: Comprehensive Google Workspace permissions
- **Token Storage**: File-based with automatic refresh
- **Credentials**: JSON file from Google Cloud Console

### Google API Scopes
#### Gmail Scopes
- `gmail.GmailLabelsScope` - Label management
- `gmail.GmailModifyScope` - Email modification
- `gmail.MailGoogleComScope` - Full Gmail access
- `gmail.GmailSettingsBasicScope` - Basic settings

#### Calendar Scopes  
- `calendar.CalendarScope` - Full calendar access
- `calendar.CalendarEventsScope` - Event management

#### Chat Scopes (Extensive)
- Admin memberships and spaces (read/write)
- Messages (create, read, reactions)
- User read states
- Space management

### Security Features
- **Token Encryption**: Automatic OAuth token refresh
- **Scope Limitation**: Granular permission control
- **Credential Isolation**: Separate credentials and token files
- **Security Scanning**: TruffleHog integration for secret detection

## Deployment Considerations

### Binary Distribution
- **Static Linking**: CGO disabled for portable binaries
- **Size Optimization**: Build flags `-ldflags="-s -w"` for smaller binaries
- **Cross Compilation**: Go's built-in cross-platform support

### Integration Requirements
- MCP-compatible AI system (Claude, etc.)
- Google Cloud Platform project with enabled APIs
- Valid OAuth 2.0 credentials for Google services
- File system access for credential and token storage

### Performance Characteristics
- **Concurrency**: Go's goroutines for efficient request handling
- **Memory Usage**: Minimal overhead with modular tool loading
- **Startup Time**: Fast initialization with lazy service creation
- **Network**: HTTP/2 support via Google API clients