# One Go - Project TODO

## 📋 Analisis Repository OmniRoute (diegosouzapw/OmniRoute)

### 1. Database Tables (SQLite)

| Table | Deskripsi |
|-------|-----------|
| `provider_connections` | Provider OAuth/API key registration dan credentials |
| `models` | Model definitions, capabilities, pricing |
| `combos` | Combo routing configs |
| `combo_targets` | Target ordering untuk combos |
| `api_keys` | API key lifecycle, scopes, quota tracking |
| `settings` | KV store untuk system configuration |
| `secrets` | Encrypted secret storage |
| `usage_history` | Request usage history |
| `call_logs` | Call/request logs |
| `proxy_logs` | Proxy usage logs |
| `proxy_registry` | Proxy configurations |
| `proxy_assignments` | Provider-proxy assignments |
| `quota_snapshots` | Historical quota usage |
| `quota_pools` | Quota-Share pool management |
| `quota_allocations` | Per-key allocations |
| `quota_consumption` | Rolling counters per (apiKeyId, dimensionKey) |
| `model_combo_mappings` | Model-to-combo mappings |
| `webhooks` | Event-driven webhook subscriptions |
| `request_detail_logs` | Per-request audit logging |
| `audit_log` | Administrative action logs |
| `mcp_tool_audit` | MCP tool call audit |
| `db_meta` | Database metadata |

### 2. API Endpoints

#### Providers
- `GET /api/providers` - List all providers
- `POST /api/providers` - Create provider
- `GET /api/providers/[id]` - Get provider by ID
- `PUT /api/providers/[id]` - Update provider
- `DELETE /api/providers/[id]` - Delete provider
- `GET /api/providers/[id]/models` - Get models for provider
- `POST /api/providers/[id]/test` - Test provider connection
- `POST /api/providers/validate` - Validate provider API key

#### Combos
- `GET /api/combos` - List all combos
- `POST /api/combos` - Create combo
- `GET /api/combos/[id]` - Get combo by ID
- `PUT /api/combos/[id]` - Update combo
- `DELETE /api/combos/[id]` - Delete combo
- `GET /v1/combos` - Public combo metadata (API-key auth)

#### Models
- `GET /api/models` - Get models with aliases
- `GET /api/models/alias` - Get/Set model aliases
- `GET /api/models/catalog` - All models by provider + type

#### API Keys
- `GET /api/keys` - List API keys
- `POST /api/keys` - Create API key
- `GET /api/keys/[id]` - Get API key
- `PUT /api/keys/[id]` - Update API key
- `DELETE /api/keys/[id]` - Delete API key

#### Usage & Analytics
- `GET /api/usage/history` - Usage history
- `GET /api/usage/logs` - Usage logs
- `GET /api/usage/request-logs` - Request-level logs
- `GET /api/usage/[connectionId]` - Per-connection usage
- `GET /api/usage/quota` - Check quota
- `GET /api/usage/proxy-logs` - Proxy usage logs

#### Settings
- `GET /api/settings` - Get settings
- `PUT /api/settings` - Update settings
- `GET /api/settings/combo-defaults` - Combo global defaults

#### Proxy
- `GET /api/proxy` - List proxies
- `POST /api/proxy` - Create proxy
- `PUT /api/proxy/[id]` - Update proxy
- `DELETE /api/proxy/[id]` - Delete proxy

#### OpenAI Compatible (v1)
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Completions
- `POST /v1/embeddings` - Embeddings
- `GET /v1/models` - List models

### 3. Models & Relations

```
ProviderConnection
├── id (UUID)
├── provider (string)
├── authType (oauth/apikey/free)
├── name (string)
├── email (string)
├── apiKey (encrypted)
├── accessToken (encrypted)
├── refreshToken (encrypted)
├── isActive (bool)
├── priority (int)
├── defaultModel (string)
├── globalPriority (int)
├── rateLimitProtection (bool)
├── group (string)
├── maxConcurrent (int)
├── proxyEnabled (bool)
├── quotaWindowThresholds (JSON)
├── rateLimitOverrides (JSON)
├── providerSpecificData (JSON)
├── testStatus (string)
├── lastError (string)
├── lastTestedAt (datetime)
└── timestamps

Combo
├── id (UUID)
├── name (string)
├── strategy (priority/round-robin/weighted/fallback)
├── isHidden (bool)
├── maxRetries (int)
├── retryDelayMs (int)
├── fallbackDelayMs (int)
└── ComboTargets[] (has many)

ComboTarget
├── id (UUID)
├── comboId (FK)
├── providerId (string)
├── modelId (string)
├── priority (int)
├── weight (int)
└── connectionId (FK)

ApiKey
├── id (UUID)
├── name (string)
├── key (hashed)
├── machineId (string)
├── allowedModels (JSON array)
├── blockedModels (JSON array)
├── allowedCombos (JSON array)
├── allowedConnections (JSON array)
├── noLog (bool)
├── autoResolve (bool)
├── isActive (bool)
├── isBanned (bool)
├── expiresAt (datetime)
├── accessSchedule (JSON)
├── maxRequestsPerDay (int)
├── maxRequestsPerMinute (int)
├── throttleDelayMs (int)
├── budget (decimal)
├── usedBudget (decimal)
├── scopes (JSON array)
├── rateLimits (JSON)
└── timestamps

Proxy
├── id (UUID)
├── name (string)
├── type (http/https/socks5)
├── host (string)
├── port (int)
├── username (string)
├── password (encrypted)
├── isActive (bool)
├── isGlobal (bool)
└── timestamps

UsageHistory
├── id (auto)
├── provider (string)
├── model (string)
├── connectionId (FK)
├── apiKeyId (FK)
├── apiKeyName (string)
├── tokensInput (int)
├── tokensOutput (int)
├── tokensCacheRead (int)
├── tokensCacheCreation (int)
├── tokensReasoning (int)
├── serviceTier (string)
├── status (string)
├── success (bool)
├── latencyMs (int)
├── ttftMs (int)
├── errorCode (string)
└── timestamp

ProxyLog
├── id (auto)
├── proxyId (FK)
├── connectionId (FK)
├── status (string)
├── responseTimeMs (int)
├── error (string)
├── clientIp (string)
├── userAgent (string)
└── timestamp
```

### 4. Routing Strategies
- **priority** - Route berdasarkan prioritas (fallback ke next jika gagal)
- **round-robin** - Rotate antar providers
- **weighted** - Distribute berdasarkan weight
- **fallback** - Coba pertama, fallback jika gagal

### 5. Middleware
- JWT Authentication
- API Key Authentication
- Rate Limiting (per key, per minute, per day)
- Request Logging
- CORS
- Proxy Resolution
- Quota Enforcement

### 6. UI Pages (Dashboard)
- Dashboard (overview)
- Providers (CRUD + test)
- Combos (CRUD + mapping)
- Models (catalog + aliases)
- API Keys (CRUD + scopes)
- Usage (history + analytics)
- Logs (call logs + proxy logs)
- Proxy (CRUD + test)
- Settings (global config)

---

## ✅ Implementation Checklist

### Phase 1: Foundation [COMPLETED ✅]
- [x] Project setup dengan Go modules
- [x] Struktur folder (Clean Architecture)
- [x] Database configuration (SQLite + PostgreSQL support)
- [x] Database migrations
- [x] Basic models (User, Settings)
- [x] JWT Authentication
- [x] Login/Logout
- [x] Dashboard layout
- [x] Sidebar navigation

### Phase 2: Provider Management
- [x] Provider model + migration
- [x] Provider repository (interface + implementation)
- [x] Provider service
- [x] Provider handler (CRUD)
- [x] Provider test connection
- [x] Provider API key validation
- [x] Provider UI pages
- [x] Provider unit tests

### Phase 3: Combo (Routing)
- [x] Combo model + migration
- [x] ComboTarget model
- [x] Combo repository
- [x] Combo service
- [x] Combo handler (CRUD)
- [x] Routing strategies implementation
- [x] Combo UI pages
- [x] Combo unit tests

### Phase 4: API Keys
- [x] ApiKey model + migration
- [x] ApiKey repository
- [x] ApiKey service
- [x] ApiKey handler (CRUD)
- [x] API Key validation middleware
- [x] Rate limiting per key
- [x] Scopes & permissions
- [x] ApiKey UI pages
- [x] ApiKey unit tests

### Phase 5: Quota Management
- [x] Quota models + migrations
- [x] Quota repository
- [x] Quota service
- [x] Quota tracking
- [x] Quota enforcement middleware
- [x] Quota UI pages
- [x] Quota unit tests

### Phase 6: Usage & Logs
- [x] UsageHistory model + migration
- [x] CallLog model + migration
- [x] Usage repository
- [x] Usage service
- [x] Usage handler
- [x] Analytics aggregation
- [x] Export functionality
- [x] Usage UI pages
- [x] Log UI pages
- [x] Usage unit tests

### Phase 7: Proxy Management
- [x] Proxy model + migration
- [x] ProxyLog model + migration
- [x] Proxy repository
- [x] Proxy service
- [x] Proxy handler (CRUD)
- [x] Proxy test connection
- [x] Proxy assignment to providers
- [x] Proxy UI pages
- [x] Proxy unit tests

### Phase 8: OpenAI Compatible API
- [x] v1/chat/completions endpoint
- [x] v1/completions endpoint
- [x] v1/embeddings endpoint
- [x] v1/models endpoint
- [x] Request routing logic
- [x] Response streaming (SSE)
- [x] Error handling
- [x] Rate limiting
- [ ] API unit tests

### Phase 9: Settings & Config
- [x] Settings model + migration
- [x] Settings repository
- [x] Settings service
- [x] Settings handler
- [x] Global configurations
- [x] Settings UI pages
- [x] Settings unit tests

### Phase 10: Polish & Testing
- [ ] Integration tests
- [ ] E2E tests
- [ ] Performance optimization
- [ ] Security audit
- [ ] Documentation
- [ ] Docker support
- [ ] CI/CD pipeline

---

## 🛠️ Technical Stack

### Backend
- Go 1.24+
- Gin (HTTP framework)
- GORM (ORM)
- SQLite (default) / PostgreSQL (optional)
- JWT (authentication)
- bcrypt (password hashing)
- Zap (structured logging)
- Viper (configuration)
- Wire (dependency injection)

### Frontend
- HTMX (dynamic content)
- TailwindCSS (styling)
- Alpine.js (reactivity)
- Templ (Go templates)

### Development
- Air (hot reload)
- golangci-lint (linting)
- go test (testing)
- Docker (containerization)

---

## 📁 Folder Structure

```
one-go/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── auth.go
│   │   ├── provider.go
│   │   ├── combo.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── proxy.go
│   │   └── settings.go
│   ├── service/
│   │   ├── auth.go
│   │   ├── provider.go
│   │   ├── combo.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── proxy.go
│   │   └── settings.go
│   ├── repository/
│   │   ├── interface.go
│   │   ├── provider.go
│   │   ├── combo.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── proxy.go
│   │   └── settings.go
│   ├── model/
│   │   ├── user.go
│   │   ├── provider.go
│   │   ├── combo.go
│   │   ├── apikey.go
│   │   ├── usage.go
│   │   ├── proxy.go
│   │   └── settings.go
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── apikey.go
│   │   ├── ratelimit.go
│   │   ├── logging.go
│   │   └── cors.go
│   ├── router/
│   │   └── router.go
│   └── database/
│       ├── database.go
│       └── migrations/
├── web/
│   ├── templates/
│   │   ├── layouts/
│   │   ├── pages/
│   │   └── components/
│   └── static/
│       ├── css/
│       └── js/
├── pkg/
│   ├── logger/
│   ├── validator/
│   ├── crypto/
│   └── response/
├── migrations/
├── storage/
├── .env.example
├── go.mod
├── go.sum
├── Makefile
├── Dockerfile
└── README.md
```

---

## 📝 Notes

- Semua CRUD harus memiliki validation (menggunakan go-playground/validator)
- Semua endpoint harus memiliki unit test
- Semua query menggunakan repository pattern
- Semua konfigurasi melalui .env (menggunakan Viper)
- Semua error menggunakan custom error types
- Semua log menggunakan structured logging (Zap)
- Setiap selesai satu fitur, update TODO.md

---

*Last Updated: 2026-07-19*
