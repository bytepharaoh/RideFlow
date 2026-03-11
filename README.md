# Rideshare Backend

A production-style microservices backend for a real-time ride-sharing application.

## Services

| Service | Responsibility |
|---|---|
| API Gateway | External HTTP/WebSocket entry point |
| Trip Service | Trip lifecycle, routing, fare calculation |
| Driver Service | Driver availability, matching, offer handling |
| Payment Service | Stripe integration, payment state |

## Tech Stack

- Go — backend language
- gRPC — internal service communication
- RabbitMQ — async event-driven communication
- MongoDB — persistence
- WebSockets — real-time client updates
- Docker + Kubernetes — containerization and orchestration
- Tilt — local k8s dev loop
- Jaeger — distributed tracing
- Stripe — payments

## Getting Started

See `docs/architecture.md` for full system design.