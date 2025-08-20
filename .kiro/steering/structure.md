# Project Structure

## Root Directory Organization

### Core Files
- **main.go**: Entry point with MCP server initialization and tool registration
- **go.mod/go.sum**: Go module definition and dependency management
- **justfile**: Build automation with development commands (build, docs, scan, install)
- **.env**: Environment configuration (not tracked in git)

### Documentation & Configuration
- **README.md**: Primary project documentation with installation and usage
- **CHANGELOG.md**: Version history and release notes
- **CLAUDE.md**: Kiro spec-driven development configuration

### Build Output
- **bin/**: Compiled binaries and credential files (gitignored)
- **google-kit**: Built executable for distribution

## Subdirectory Structures

### `/tools/` - MCP Tool Implementations
Service-specific tool definitions following MCP protocol standards:
- **calendar.go**: Google Calendar operations (events, responses, listings)
- **gmail.go**: Gmail management (search, filters, labels, spam handling) 
- **gchat.go**: Google Chat functionality (messaging, spaces, user management)

**Pattern**: Each file registers multiple related tools via `Register[Service]Tools(server)` function

### `/services/` - Google API Integration
Core service layer for Google API client management:
- **google.go**: OAuth2 authentication, scope management, HTTP client creation
- **gchat.go**: Google Chat service-specific implementations  
- **httpclient.go**: HTTP client configuration and proxy support

**Pattern**: Service initialization with OAuth token management and scope configuration

### `/util/` - Common Utilities
Shared functionality across the codebase:
- **handler.go**: Error handling wrappers with panic recovery and stack traces

**Pattern**: Utility functions that enhance MCP tool reliability and debugging

### `/scripts/` - Development & Setup Tools
Helper scripts for project setup and maintenance:
- **get-google-token/**: OAuth token generation utility with detailed README
  - **main.go**: Interactive token generation program
  - **README.md**: Step-by-step Google Cloud Platform setup guide
- **docs/**: Documentation generation utilities
  - **update-doc.go**: Automated documentation updater

**Pattern**: Self-contained tools with their own documentation and main functions

### `/.kiro/steering/` - Spec-Driven Development
Generated steering documents for development guidance:
- **product.md**: Product overview and business context
- **tech.md**: Technology stack and architectural decisions
- **structure.md**: This file - code organization patterns

## Code Organization Patterns

### Modular Tool Registration
```go
// main.go pattern
if isEnabled("calendar") {
    tools.RegisterCalendarTools(mcpServer)
}
```
**Principle**: Selective loading based on environment configuration

### Service Layer Architecture
```go
// services/google.go pattern
func GoogleHttpClient(tokenFile, credentialsFile string) *http.Client {
    // OAuth2 configuration and client creation
}
```
**Principle**: Centralized authentication and API client management

### Error Handling Strategy
```go
// util/handler.go pattern
func ErrorGuard(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
    // Panic recovery with stack traces
}
```
**Principle**: Robust error handling with detailed debugging information

### MCP Tool Definition
```go
// tools/*.go pattern
eventTool := mcp.NewTool("calendar_event",
    mcp.WithDescription("Manage Google Calendar events"),
    mcp.WithString("action", mcp.Required(), ...),
)
```
**Principle**: Declarative tool definitions with comprehensive parameter specifications

## File Naming Conventions

### Package Structure
- **Single-word packages**: `tools`, `services`, `util` (lowercase, no underscores)
- **Service-based files**: Named after Google service (calendar.go, gmail.go, gchat.go)
- **Function-based files**: Named after primary purpose (handler.go, httpclient.go)

### Function Naming
- **Public functions**: PascalCase with clear service prefixes (`RegisterCalendarTools`)
- **Private functions**: camelCase with descriptive names (`isEnabled`, `tokenFromFile`)
- **Tool names**: snake_case matching MCP conventions (`calendar_event`, `gmail_search`)

### Variable Naming
- **Environment variables**: SCREAMING_SNAKE_CASE (`GOOGLE_CREDENTIALS_FILE`)
- **Local variables**: camelCase (`mcpServer`, `enableTools`)
- **Constants**: PascalCase for exported, camelCase for private

## Import Organization

### Standard Library First
```go
import (
    "context"
    "fmt" 
    "os"
    // ... other standard library
)
```

### Third-party Dependencies
```go
import (
    "github.com/joho/godotenv"
    "github.com/mark3labs/mcp-go/mcp"
    "golang.org/x/oauth2"
    "google.golang.org/api/calendar/v3"
)
```

### Local Imports Last
```go
import (
    "github.com/nguyenvanduocit/google-kit/services"
    "github.com/nguyenvanduocit/google-kit/util"
)
```

## Key Architectural Principles

### Separation of Concerns
- **Tools**: MCP protocol implementation and parameter handling
- **Services**: Google API integration and authentication
- **Utilities**: Cross-cutting concerns (error handling, logging)

### Configuration-Driven Behavior
- Environment variables control tool loading
- File-based credential management
- Modular service enabling via `ENABLE_TOOLS`

### Error Resilience
- Panic recovery at tool handler level
- Detailed error reporting with stack traces
- Graceful degradation for missing services

### Extensibility
- Plugin-like tool registration pattern
- Service-agnostic utility functions
- Clear interfaces between layers

### Security-First Design
- Credential isolation in separate files
- OAuth2 best practices with token refresh
- No hardcoded secrets or credentials in source code

## Development Workflow Integration

### Build Process
- **Static binaries**: CGO disabled for portability
- **Optimization**: Linker flags for size reduction (`-ldflags="-s -w"`)
- **Just commands**: Standardized development tasks

### Documentation Maintenance
- Automated documentation generation via `scripts/docs/`
- Comprehensive README with setup instructions
- Inline code documentation following Go conventions

### Security Scanning
- TruffleHog integration for secret detection
- Clean separation of credentials from source code
- Regular security scanning via `just scan` command