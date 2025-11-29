Got it da — you want the **same README**, but fully rewritten to match your **Collaborative Workspace Platform**, NOT a file storage platform.
I rewrote everything with correct terminology: **Workspaces, Channels, Groups, Chat, File Sharing, Real-time Collaboration, Activity Feed**, etc.

Here is the corrected, polished, company-grade README for your **Collaborative Platform Backend**.

---

# 🚀 **Project Name: Collab Hub Backend**

## 🌟 1. Overview

This is the backend repository for **Collab Hub**, a modern collaborative workspace platform designed for teams, communities, and organizations.
It brings together **group chat**, **direct messaging**, **shared workspaces**, **file sharing**, **real-time updates**, and **team collaboration tools**, powered by a clean and scalable backend architecture.

Built using:

* **Go (Golang)**
* **Microservices**
* **gRPC communication**
* **Clean Architecture**
* **PostgreSQL + Redis**
* **S3-compatible storage (MinIO / AWS S3)**
* **Docker & Docker Compose**

This architecture ensures isolated, independent services with clear domain boundaries, allowing massive scalability and clean development workflows.

---

## 🧱 2. Core Architecture & Services

Your platform is composed of several independently deployable services.

| Service                  | Purpose                                                                                                 | Storage                                  |
| ------------------------ | ------------------------------------------------------------------------------------------------------- | ---------------------------------------- |
| **auth-service**         | User registration, login, OTP/Magic Link verification, sessions, tokens.                                | PostgreSQL (`auth_db`)                   |
| **user-service**         | User profiles, settings, presence, status.                                                              | PostgreSQL (`user_db`)                   |
| **workspace-service**    | Workspace management, teams, channels, roles, workspace membership.                                     | PostgreSQL (`workspace_db`)              |
| **chat-service**         | Real-time group chat, direct messages, message history, typing indicators.                              | PostgreSQL (`chat_db`) + Redis (Pub/Sub) |
| **file-service**         | File uploads, previews, attachments in chat, workspace file storage.                                    | PostgreSQL (`file_db`) + MinIO/S3        |
| **notification-service** | Push notifications, email alerts, workspace activity notifications.                                     | PostgreSQL (`notification_db`)           |
| **api-gateway**          | Public HTTP REST entrypoint → Converts HTTP → gRPC → Returns JSON responses to frontend/mobile clients. | Stateless                                |
| **cron-service**         | Background tasks like file cleanup, offline user marking, workspace cleanup.                            | None                                     |

---

## 📁 3. Directory Structure

The backend follows **Clean Architecture + Modular Microservices.**

| Folder                   | Description                                                       |
| ------------------------ | ----------------------------------------------------------------- |
| `api/proto/v1`           | All `.proto` contracts for every service.                         |
| `cmd/`                   | Each service's `main.go` entrypoint.                              |
| `configs/`               | Service-specific YAML configuration files.                        |
| `internal/`              | All business logic (not importable by others).                    |
| `internal/chat/app/`     | Chat use cases: send message, broadcast, history, attachments.    |
| `internal/chat/adapter/` | PostgreSQL repositories, Redis pub/sub handlers, gRPC servers.    |
| `internal/chat/port/`    | Interfaces for Chat service ports (repositories, event bus, etc). |
| `migrations/`            | SQL migrations per service.                                       |
| `pkg/`                   | Shared utilities: logger, config loader, database tools.          |
| `pkg/pb/`                | Auto-generated gRPC Go code from `.proto` files.                  |

---

## 🔗 4. Chat Service Flow (Core of Your Platform)

Your real-time chat is powered by:

```
Client → WebSocket → chat-service → Redis PubSub → DB → WebSocket → All clients
```

**Detailed Flow:**

1. Client connects via WebSocket to **chat-service**.
2. User sends message → chat-service validates session.
3. Message is saved in PostgreSQL (chat DB).
4. Message is published via Redis → `channel:{workspaceId}` or `dm:{userId}`.
5. All subscribed chat-service instances receive the message from Redis.
6. chat-service pushes it to all connected clients via WebSocket.
7. Delivery receipts, typing indicators, attachments flow across same pipeline.

---

## 🛠️ 5. Local Setup & Development

### Prerequisites

* Go 1.20+
* Docker & Docker Compose
* Protocol Buffers (`protoc`)
* `buf` or `protoc-gen-go`

---

### Step 1: Clone

```bash
git clone https://github.com/ak-repo/collab-hub
cd collab-hub/backend
go mod tidy
```

### Step 2: Start Infrastructure

```bash
docker compose up -d
```

This starts:

* 4 PostgreSQL DBs
* Redis
* MinIO
* All services (optional)

---

### Step 3: Generate gRPC Code

```bash
make proto
```

---

### Step 4: Run Migrations

```bash
make migrate-auth
make migrate-chat
make migrate-workspace
make migrate-files
```

---

### Step 5: Run Services

```bash
go run ./cmd/auth-service/main.go
go run ./cmd/chat-service/main.go
go run ./cmd/api-gateway/main.go
```

---

## 🧪 6. Testing

* **Unit tests** for all use cases.
* **Mocked repository tests** for each service.
* **Integration tests** via full gRPC communication.

Run tests:

```bash
go test ./...
```

---

## 💡 7. Contribution

* PRs must follow the architecture.
* No business logic in `adapter`.
* Only interface types in `port/`.
* `app/` contains actual use case logic.

---

If you want, I can also generate:

✅ `Makefile`
✅ full folder skeleton
✅ docker-compose.yml for your entire system
✅ chat-service full implementation
✅ workspace-service boilerplate
✅ WebSocket gateway code

Just tell me.
