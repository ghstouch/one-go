# OmniRoute Go

AI Gateway/Router built with Go - A self-hosted LLM routing solution.

## Features

- 🔌 **Provider Management** - Connect multiple AI providers (OpenAI, Claude, etc.)
- 🎯 **Combo Routing** - Configure routing strategies (priority, round-robin, weighted, fallback)
- 📊 **Usage Tracking** - Monitor requests, tokens, and costs
- 🔑 **API Key Management** - Generate and manage API keys with scopes
- 🌐 **Proxy Support** - HTTP/SOCKS5 proxy configuration
- 🔐 **JWT Authentication** - Secure dashboard access
- 📝 **Request Logging** - Detailed request/response logging

## Tech Stack

- **Backend**: Go 1.24+, Gin, GORM
- **Database**: SQLite (default), PostgreSQL (optional)
- **Frontend**: HTMX, TailwindCSS, Alpine.js
- **Auth**: JWT

## Quick Start

### Prerequisites

- Go 1.24 or later
- Make (optional, for using Makefile commands)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/omniroute-go.git
cd omniroute-go
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Edit `.env` with your configuration:
```env
# Server
SERVER_PORT=8080
GIN_MODE=debug

# Database
DB_DRIVER=sqlite
SQLITE_PATH=./storage/omniroute.db

# JWT
JWT_SECRET=your-secret-key

# Admin
ADMIN_USERNAME=admin
ADMIN_PASSWORD=admin123
```

4. Run the application:
```bash
# Using Make
make run

# Or directly with Go
go run ./cmd/server/main.go
```

5. Open your browser at `http://localhost:8080`

### Default Login

- Username: `admin`
- Password: `admin123`

## API Endpoints

### Authentication
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token
- `GET /api/auth/me` - Get current user info

### Dashboard
- `GET /api/dashboard/stats` - Get dashboard statistics
- `GET /api/dashboard/activity` - Get recent activity

### Settings
- `GET /api/settings` - Get all settings
- `PUT /api/settings` - Update a setting

### Health
- `GET /api/health` - Health check endpoint

### Providers
- `GET /api/providers` - List all providers
- `POST /api/providers` - Create provider
- `PUT /api/providers/:id` - Update provider
- `DELETE /api/providers/:id` - Delete provider
- `POST /api/providers/:id/test` - Test provider connection
- `POST /api/providers/validate` - Validate provider API key

### Combos
- `GET /api/combos` - List all combos
- `POST /api/combos` - Create combo
- `GET /api/combos/:id` - Get combo by ID
- `PUT /api/combos/:id` - Update combo
- `DELETE /api/combos/:id` - Delete combo

### API Keys
- `GET /api/keys` - List API keys
- `POST /api/keys` - Generate API key
- `GET /api/keys/:id` - Get API key
- `PUT /api/keys/:id` - Update API key
- `DELETE /api/keys/:id` - Delete API key

### Usage & Logs
- `GET /api/usage` - List usage history (paginated, filterable)
- `GET /api/usage/stats` - Get usage statistics
- `GET /api/usage/logs` - List call logs
- `GET /api/usage/logs/:id` - Get call log detail
- `GET /api/usage/export` - Export usage as CSV

### Quota
- `GET /api/quota` - Get all quotas
- `GET /api/quota/:provider` - Get provider quota

### Proxy
- `GET /api/proxy` - List proxies
- `POST /api/proxy` - Create proxy
- `PUT /api/proxy/:id` - Update proxy
- `DELETE /api/proxy/:id` - Delete proxy
- `POST /api/proxy/:id/test` - Test proxy connection
- `GET /api/proxy/logs` - Get proxy logs

### Webhooks
- `GET /api/webhooks` - List webhooks
- `POST /api/webhooks` - Create webhook
- `PUT /api/webhooks/:id` - Update webhook
- `DELETE /api/webhooks/:id` - Delete webhook
- `GET /api/webhooks/logs` - Get webhook delivery logs

### Audit Logs
- `GET /api/audit-logs` - List audit logs (paginated, filterable)
- `GET /api/audit-logs/:id` - Get audit log detail

### Model Aliases
- `GET /api/models/aliases` - List model aliases
- `POST /api/models/aliases` - Create model alias
- `DELETE /api/models/aliases/:id` - Delete model alias
- `GET /api/models/resolve` - Resolve alias to actual model

### OpenAI-Compatible v1 API (requires API key)
- `POST /v1/chat/completions` - Chat completions (supports streaming)
- `POST /v1/completions` - Text completions
- `POST /v1/embeddings` - Text embeddings
- `GET /v1/models` - List available models

## Project Structure

```
omniroute-go/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection & migrations
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Middleware (auth, logging, cors, etc.)
│   ├── model/                # Data models
│   ├── repository/           # Data access layer
│   ├── router/               # Route definitions
│   └── service/              # Business logic
├── pkg/
│   ├── logger/               # Structured logging
│   └── response/             # Response helpers
├── web/
│   ├── static/               # Static files (CSS, JS)
│   └── templates/            # HTML templates
├── storage/                  # SQLite database files
├── .env                      # Environment configuration
├── go.mod                    # Go modules
├── Makefile                  # Build commands
└── README.md                 # This file
```

## Development

### With Hot Reload

Install Air for hot reload:
```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:
```bash
make dev
```

### Run Tests

```bash
make test
```

### Run Linter

```bash
make lint
```

## Docker

### Using Docker Compose (recommended)

```bash
docker-compose up -d
```

### Build Image

```bash
make docker-build
```

### Run Container

```bash
make docker-run
```

## License

MIT License

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request
