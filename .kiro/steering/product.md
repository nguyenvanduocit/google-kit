# Product Overview

## Product Overview

Google Kit is a Model Context Protocol (MCP) server implementation that bridges AI models with Google Workspace services. It enables AI assistants like Claude to interact with Google Calendar, Gmail, and Google Chat through a standardized interface, allowing for seamless automation and integration of Google services in AI workflows.

## Core Features

- **Google Calendar Integration**: Create, list, update events and respond to invitations
- **Gmail Management**: Search emails, manage labels and filters, move spam, and organize inbox
- **Google Chat Operations**: Send messages, list spaces, and manage chat interactions
- **Flexible Tool Configuration**: Selective enabling of tool groups (calendar, gmail, gchat)
- **OAuth 2.0 Authentication**: Secure Google API access with proper credential management
- **MCP Standard Compliance**: Compatible with any MCP-enabled AI system
- **Environment-based Configuration**: Easy deployment with environment variable setup

## Target Use Case

### Primary Scenarios
- **AI-Powered Productivity**: Enable AI assistants to manage calendars, emails, and chat on behalf of users
- **Workflow Automation**: Integrate Google services into larger AI-driven automation pipelines
- **Enterprise Integration**: Connect Google Workspace with AI tools for enhanced productivity
- **Personal Assistant Applications**: Build AI assistants that can interact with personal Google accounts

### Specific Applications
- Schedule management through natural language
- Email organization and filtering automation
- Cross-platform communication via Google Chat
- Event planning and coordination
- Inbox management and email processing

## Key Value Proposition

### Unique Benefits
- **Standardized Integration**: Uses MCP protocol for consistent AI-to-service communication
- **Granular Control**: Selective tool enabling allows customized deployment scenarios
- **Security-First**: Proper OAuth 2.0 implementation with secure token management
- **Extensible Architecture**: Modular design allows easy addition of new Google services
- **Cross-Platform Compatibility**: Works with any MCP-compatible AI system, not just specific providers

### Differentiators
- **MCP Native**: Built specifically for the Model Context Protocol standard
- **Google-Focused**: Deep integration with Google Workspace ecosystem
- **Production Ready**: Includes proper error handling, logging, and configuration management
- **Developer Friendly**: Clear documentation and modular codebase for easy contribution

## Business Context

### Project Status
- **Current Version**: 1.0.0
- **Development Stage**: Production-ready initial release
- **Maintenance Model**: Open source project with active development
- **License**: Open source (check repository for specific license)

### Integration Requirements
- Google Cloud Platform account with enabled APIs
- OAuth 2.0 credentials configuration  
- MCP-compatible AI system (Claude, etc.)
- Go runtime environment for deployment