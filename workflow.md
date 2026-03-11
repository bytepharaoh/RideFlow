rideshare/
├── cmd/                        # Entry points — one folder per runnable service
│   ├── gateway/
│   │   └── main.go
│   ├── trip/
│   │   └── main.go
│   ├── driver/
│   │   └── main.go
│   └── payment/
│       └── main.go
│
├── internal/                   # Private application code — NOT importable by outside modules
│   ├── gateway/                # Gateway-specific logic
│   │   ├── handler/
│   │   ├── middleware/
│   │   └── ws/
│   ├── trip/                   # Trip service domain
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── repository/
│   │   └── service/
│   ├── driver/                 # Driver service domain
│   │   ├── domain/
│   │   ├── handler/
│   │   ├── repository/
│   │   └── service/
│   └── payment/                # Payment service domain
│       ├── domain/
│       ├── handler/
│       ├── repository/
│       └── service/
│
├── pkg/                        # Shared code that IS safe to reuse across services
│   ├── logger/                 # Structured logging setup
│   ├── config/                 # Config loading helpers
│   ├── messaging/              # RabbitMQ client wrapper
│   ├── middleware/             # Shared HTTP middleware (e.g. tracing headers)
│   └── grpcutil/               # Shared gRPC helpers (interceptors, etc.)
│
├── api/                        # API contracts (OpenAPI specs, HTTP schemas)
│   └── openapi/
│
├── proto/                      # Protobuf definitions (source of truth for gRPC)
│   ├── trip/
│   │   └── trip.proto
│   ├── driver/
│   │   └── driver.proto
│   └── payment/
│       └── payment.proto
│
├── deployments/                # All deployment config
│   ├── docker/
│   │   ├── gateway.Dockerfile
│   │   ├── trip.Dockerfile
│   │   ├── driver.Dockerfile
│   │   └── payment.Dockerfile
│   ├── k8s/                    # Kubernetes manifests
│   │   ├── gateway/
│   │   ├── trip/
│   │   ├── driver/
│   │   └── payment/
│   └── tilt/
│       └── Tiltfile
│
├── scripts/                    # Shell scripts for dev tasks
│   ├── proto-gen.sh
│   └── seed.sh
│
├── docs/                       # Architecture docs, ADRs, diagrams
│   └── architecture.md
│
├── .env.example                # Example env vars — NEVER commit real secrets
├── .gitignore
├── docker-compose.yml          # For local infra (RabbitMQ, MongoDB, Jaeger)
├── go.mod
├── go.sum
├── Makefile
└── README.md