┌─────────────────────────────────────────────────────┐
│            PhoenixUpime  Backend (phoenix-go)       │
│           github.com/Phoenix-Uptime/phoenix-go      │
│                                                     │
│   ┌───────────────────────────────────────────────┐ │
│   │            Fiber Web Framework                │ │
│   │ (Handles HTTP requests and routing)           │ │
│   └───────────────┬───────────────────────────────┘ │
│                   │                                 │
│                   │                                 │
│    ┌──────────────▼──────────────┐                  │
│    │   Authentication Layer      │                  │
│    │  (Session ID / API Key)     │                  │
│    └───────────────┬─────────────┘                  │
│                    │                                │
│   ┌────────────────▼────────────────┐               │
│   │          API Endpoints          │               │
│   │   (User, Monitor, Alert APIs)   │               │
│   └────────────────┬────────────────┘               │
│                    │                                │
│ ┌──────────────────▼──────────────────┐             │
│ │      Request Validation & Parsing   │             │
│ │   (Validates input, parses JSON)    │             │
│ └──────────────────┬──────────────────┘             │
│                    │                                │
│   ┌────────────────▼────────────────┐               │
│   │  Monitoring Service (Scheduler) │               │
│   │  ─ Performs regular uptime checks│              │
│   │  ─ Logs performance data         │              │
│   └────────────────┬─────────────────┘              │
│                    │                                │
│    ┌───────────────▼───────────────┐                │
│    │  Uptime Check Executors       │                │
│    │  ─ Sends HTTP requests        │                │
│    │  ─ Measures response times    │                │
│    └───────────────┬───────────────┘                │
│                    │                                │
│   ┌────────────────▼────────────────┐               │
│   │   Result Processing             │               │
│   │   ─ Analyzes response status    │               │
│   │   ─ Determines alert triggers   │               │
│   └────────────────┬────────────────┘               │
│                    │                                │
│    ┌───────────────▼───────────────┐                │
│    │     Alerting Service          │                │
│    │  ─ Sends notifications        │                │
│    │  ─ Email, SMS, Webhooks       │                │
│    └───────────────┬───────────────┘                │
│                    │                                │
│    ┌───────────────▼───────────────┐                │
│    │     Database Layer             │               │
│    │  (GORM - Postgres/SQLite)      │               │
│    └───────────────┬───────────────┘                │
│                    │                                │
│    ┌───────────────▼───────────────┐                │
│    │   BadgerDB Caching Layer      │                │
│    │  ─ Caches session data        │                │
│    │  ─ Stores temporary results   │                │
│    └───────────────────────────────┘                │
└─────────────────────────────────────────────────────┘

