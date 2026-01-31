# 6. Components

## 6.1 Component Overview

```mermaid
graph TB
    subgraph "CLI Application"
        CLI[CLI Layer<br/>Cobra Commands]
        TUI[TUI Layer<br/>Bubble Tea Models]
        Config[Config Manager<br/>Viper]
        HTTPClient[HTTP Client<br/>Resty]
        Auth[Auth Store<br/>Token Management]
    end
    
    subgraph "API Server"
        Router[Router<br/>Gin]
        Middleware[Middleware<br/>Auth, CORS, Logging]
        Handlers[Handlers<br/>Request Processing]
        Services[Service Layer<br/>Business Logic]
        Repos[Repository Layer<br/>Data Access]
        Email[Email Service<br/>SMTP/SendGrid]
    end
    
    subgraph "Data"
        DB[(SQLite)]
    end
    
    CLI --> Config
    CLI --> HTTPClient
    CLI --> Auth
    TUI --> HTTPClient
    HTTPClient --> Router
    Router --> Middleware
    Middleware --> Handlers
    Handlers --> Services
    Services --> Repos
    Services --> Email
    Repos --> DB
```

## 6.2 Component Responsibilities

| Component | Responsibility |
|-----------|----------------|
| **CLI Layer** | Parse commands, flags; orchestrate TUI and API calls |
| **TUI Layer** | Render interactive terminal interfaces (rooms, wizard, dashboard) |
| **Config Manager** | Load/save configuration from multiple sources |
| **HTTP Client** | Make authenticated API calls with retry logic |
| **Auth Store** | Secure token storage (OS keychain) |
| **API Router** | Route HTTP requests, apply middleware |
| **Middleware** | Auth, CORS, logging, recovery |
| **Handlers** | Request validation, response formatting |
| **Service Layer** | Business logic (BookingService, RoomService, etc.) |
| **Repository Layer** | Data access abstraction |
| **Email Service** | Template rendering, SMTP delivery |

---
