# SVIX Webhook Service

A Go-based webhook service that processes Google Pub/Sub messages and forwards them to Svix webhooks.
## Architecture

The service follows a clean architecture pattern with the following key components:

### Core Components

1. **Svix Client** (`internal/svix/client.go`)
   - Handles communication with Svix API
   - Implements retry logic for failed requests
   - Manages application creation and message sending

2. **Event Handler** (`internal/handler/handler.go`)
   - Processes incoming Pub/Sub messages
   - Validates and transforms event data
   - Routes events to appropriate Svix endpoints

3. **Models** (`internal/models/`)
   - `BaseEvent`: Core event structure with ID, Type, and Data
   - `PubSubMessage`: Google Pub/Sub message structure
   - Additional event-specific models
  
### Key Interfaces

// Client interface for Svix operations
```
type Client interface {
    CreateApplication(ctx context.Context, name string) (string, error)
    SendMessage(ctx context.Context, appID string, event models.BaseEvent) error
}
```
// Handler interface for processing events
```
type Handler interface {
    ProcessEvent(c gin.Context)
}
```

## Project Structure
```
gigs-challenge/
├── cmd/
│   └── webhook-service/
│       └── main.go           # Application entry point
├── internal/
│   ├── controllers/
│   │   ├── notification.go   # HTTP request handling
│   │   └── parser.go        # Request parsing logic
│   ├── logger/
│   │   └── logger.go        # Logging configuration
│   ├── models/
│   │   ├── event.go         # Event data structures
│   │   ├── pubsub.go        # Pub/Sub message structures
│   │   └── webhook.go       # Webhook event types
│   ├── router/
│   │   └── router.go        # HTTP routing setup
│   ├── svix/
│   │   ├── client.go        # Svix client implementation
│   │   ├── init.go          # Application initialization
│   │   └── retry.go         # Retry logic
│   ├── utils/
│   │   └── error.go         # Error handling utilities
│   ├── webhooks/
│   │   └── webhook.go       # Webhook processing
│   └── worker/
│       └── pool.go          # Worker pool implementation
├── test/
│   ├── events/             # Test event JSON files
│   └── run.sh             # Test runner script
├── scripts/
│   └── delete-svix-apps.sh # Cleanup utility
├── .air.toml              # Air configuration for hot reload
├── .env                   # Environment variables
├── .gitignore
├── .golangci.yml         # Linter configuration
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile             # Build and development commands
└── README.md            # Project documentation
```

### First Time Setup
```bash
# Install golangci-lint (required only once)
make install-lint
```

After installation, you can run the linter and other commands:

```bash
# Run linter
make lint

# Run tests
make test

# Build the service
make build

# Run the service
make run
```
## Development

Common commands:

### Run tests
```
make test
```

### Build the service
```
make build
```

### Run the service
```
make run
```

## Configuration

Set the following environment variables:
```
export SVIX_AUTH_TOKEN=your_svix_token
export PORT=8080
```
## API Endpoints

### POST /notifications

Receives Pub/Sub events and forwards them to Svix.

**Request Body:**
```
{
    "message": {
        "data": "base64-encoded-event-data",
        "messageId": "unique-message-id",
        "publishTime": "2024-03-21T12:00:00Z"
    }
}
```
For more information about design decisions and future improvements, see [NOTES.md](NOTES.md).

## Code Formatting

The project uses multiple layers of code formatting:

### 1. VS Code Settings (Real-time/On Save)
- Formats as you code
- Handles individual files you're working on
- Configure in VS Code settings:
```json
{
    "editor.formatOnSave": true,
    "go.formatTool": "goimports",
    "[go]": {
        "editor.defaultFormatter": "golang.go",
        "editor.formatOnSave": true,
        "editor.codeActionsOnSave": {
            "source.organizeImports": "explicit"
        }
    },
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.useLanguageServer": true
}
```

### 2. Makefile Command (Manual/Bulk)
```
bash
make fmt
```

- Formats ALL files in the project
- Uses gofmt and goimports
- Useful for bulk formatting
- Good for CI/CD pipelines

### 3. Linter Verification (.golangci.yml)
- Verifies formatting with golangci-lint
- Reports formatting violations
- Run with:
```
bash
make lint
```


### Google Pub/Sub

When Google Cloud Pub/Sub delivers a message via HTTP push, it wraps the actual message in an envelope. The structure looks like this:
```json
{
    "message": {
        "data": "eyJpZCI6IjEyMzQiLCJ0eXBlIjoic3Vic2NyaXB0aW9uLmFjdGl2YXRlZCIsImNyZWF0ZWRfYXQiOiIyMDI0LTAzLTI3VDEyOjAwOjAwWiIsImRhdGEiOnsiZm9vIjoiYmFyIn19",
        "messageId": "123456",
        "publishTime": "2024-03-27T12:00:00.000Z"
    },
    "subscription": "projects/myproject/subscriptions/mysubscription"
}
```

Where:
1. The outer structure is the Pub/Sub envelope
2. The message.data field contains our base64-encoded event
3. When decoded, the data contains our BaseEvent:

```
{
    "id": "1234",
    "type": "subscription.activated",
    "created_at": "2024-03-27T12:00:00Z",
    "data": {
        "foo": "bar"
    }
}
```

That's why we have this flow in the receiver:
1. Parse Pub/Sub envelope
2. Base64 decode the data field
3. Parse the decoded data into our BaseEvent

Looking at the Gigs Projects documentation, let me clarify the hierarchy:
```
Organization
    └── Projects
         └── Users
             └── Subscriptions, Plans, Devices, SIMs
```

Key points:

1. Organizations can have multiple projects
2. Each project has:
- Unique ID (e.g., "gigs")
- Organization reference
- Configuration for billing, payments, etc.
- Users and their subscriptions
  
This means for our webhook service:

We should create one Svix application per project, not per organization or user, because:
- Projects are the main isolation boundary
- All user/subscription data belongs to a project
- Projects have their own configuration and settings
  
This ensures events are properly isolated per project and customers (project owners) receive only their relevant events. 


## Svix Cleanup Script

This utility script helps you clean up all Svix applications in your environment.
1. Install the [Svix CLI] 
```
brew install svix/svix/svix
```
2. Make sure you have your Svix authentication token ready


### Usage

1. Make the script executable (first time only):
   ```bash
   chmod +x scripts/delete-svix-apps.sh
   ```

2. Run the script:
   ```bash
   SVIX_AUTH_TOKEN=${SVIX_AUTH_TOKEN} ./scripts/delete-svix-apps.sh
   ```

### What it does

- Lists all Svix applications in your account
- Automatically deletes each application (no confirmation required)
- Provides feedback on the deletion process

## Test Runner Script

This utility script helps test the webhook service by sending sample Pub/Sub messages to your local environment.

### Prerequisites

1. Make sure the webhook service is running locally
2. The service should be listening on port 8080 (default)

### Usage

1. Make the script executable (first time only):
   ```bash
   chmod +x test/run.sh
   ```

2. Run the script:
   ```bash
   ./test/run.sh http://localhost:8080/notifications
   ```


### What it does

- Sends sample Pub/Sub messages to your webhook endpoint
- Each message simulates different event types
- Messages include base64-encoded payloads matching the expected format
- Useful for local development and testing