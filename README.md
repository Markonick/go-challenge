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
│ └── webhook-service/
│ └── main.go # Application entry point and coordination
├── internal/
│ ├── models/
│ │ ├── event.go # Pub/Sub message structures
│ │ └── webhook.go # Svix message structures
│ ├── receiver/ # Pub/Sub event handling
│ │ ├── receiver.go # HTTP handler + event receiving logic
│ │ └── receiver_test.go
│ └── sender/ # Svix webhook delivery
│ ├── sender.go
│ └── sender_test.go
├── config/
│ └── config.go # Configuration management
├── go.mod # Go module definition
├── go.sum # Go module dependencies
├── CHALLENGE.md # Original challenge description
├── NOTES.md # Design decisions and future improvements
└── README.md # Project overview and setup instructions
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
export SVIX_API_KEY=your_api_key
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