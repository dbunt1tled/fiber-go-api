# Fiber Go Boilerplate

![FiberAPIgo](assets/images/fiber_go.png)

[![Go Version](https://img.shields.io/github/go-mod/go-version/dbunt1tled/fiber-go-api)](https://golang.org/)
[![Go Reference](https://pkg.go.dev/badge/github.com/dbunt1tled/fiber-go-api.svg)](https://pkg.go.dev/github.com/dbunt1tled/fiber-go-api)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/dbunt1tled/fiber-go-api)](https://goreportcard.com/report/github.com/dbunt1tled/fiber-go-api)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/dbunt1tled/fiber-go-api)](https://github.com/dbunt1tled/fiber-go-api/releases)
[![Build Status](https://github.com/dbunt1tled/fiber-go-api/actions/workflows/release.yml/badge.svg)](https://github.com/dbunt1tled/fiber-go-api/actions/workflows/release.yml)

A modern, modular Go web application boilerplate built with [Fiber v3](https://docs.gofiber.io/). It features a repository pattern, structured logging, background task processing, and a robust validation system.

## 🚀 Features

- **Framework**: [Fiber v3](https://docs.gofiber.io/) for high-performance web routing.
- **Architecture**: Modular design with a clean separation of concerns (Controller, Service, Repository).
- **Database**: 
  - [PostgreSQL](https://www.postgresql.org/) with [pgxpool](https://github.com/jackc/pgx) for connection pooling.
  - [goqu](https://github.com/doug-martin/goqu) for type-safe query building.
  - [goose](https://github.com/pressly/goose) for database migrations.
- **Background Tasks**: [Asynq](https://github.com/hibiken/asynq) (Redis-based) for asynchronous job processing (e.g., sending emails).
- **Security**: 
  - JWT-based authentication.
  - Password hashing with Argon2/Bcrypt (via custom hasher).
- **Validation**: [validator/v10](https://github.com/go-playground/validator) with custom validators (e.g., database uniqueness checks).
- **Logging**: Structured logging with `slog`, featuring a pretty-printed handler for development.
- **Configuration**: Environment-based configuration using `koanf`.
- **Mailing**: SMTP-based mailer with support for both synchronous and asynchronous sending.
- **Views**: HTML templates support via Fiber's template engine.

## 🛠 Tech Stack

- **Go**: 1.25.0+
- **Database**: PostgreSQL
- **Cache/Queue**: Redis
- **Web Framework**: Fiber v3
- **ORM/Query Builder**: Goqu
- **Migrations**: Goose

## 📋 Prerequisites

Before you begin, ensure you have the following installed:
- Go (1.25.0 or later)
- PostgreSQL
- Redis
- `goose` CLI (optional, but recommended for migrations)

## ⚙️ Installation & Setup

1. **Clone the repository**:
   ```bash
   git clone <repository-url>
   cd fiber_go
   ```

2. **Configure Environment Variables**:
   Copy the example environment file and update it with your settings:
   ```bash
   cp .env.example .env
   ```

3. **Install Dependencies**:
   ```bash
   go mod download
   ```

4. **Run Database Migrations**:
   Ensure your PostgreSQL instance is running and the database specified in `.env` exists.
   ```bash
   make migrate_up
   ```

## 🚀 Running the Application

### Development Mode
To run the API with hot reload (if configured) or simply via `go run`:
```bash
make run_api
```

### Build for Production
To compile the application into a binary:
```bash
make build_api
```
The binary will be located in the `bin` directory.

## 📂 Project Structure

```text
├── assets/             # Static files (CSS, Images, etc.)
├── cmd/                # Application entry points
│   └── api/           # API server entry point
├── internal/           # Private application code
│   ├── app/           # App initialization, routes, and middleware
│   ├── config/        # Configuration schema and loading logic
│   ├── lib/           # Internal libraries (email, view, etc.)
│   └── modules/       # Business logic organized by domain (auth, user)
├── migration/          # Database migration files
├── pkg/                # Public/Shared packages
│   ├── db/            # Database connection management
│   ├── hasher/        # Security and hashing utilities
│   ├── http/          # HTTP DTOs, controllers, and common middleware
│   ├── log/           # Logger implementation
│   ├── mailer/        # SMTP mailer logic
│   ├── queue/         # Redis-based background queue
│   ├── storage/       # Generic repository and filtering logic
│   └── validation/    # Request validation and custom validators
├── resources/          # Templates and other non-code resources
└── Makefile           # Build and development commands
```

## 🛠 Development Commands

Available `make` commands:

- `make run_api`: Run the API server.
- `make build_api`: Build the API binary.
- `make check_vulnerabilities`: Scan dependencies for known security issues.
- `make migrate_up`: Apply all pending database migrations.
- `make migrate_down`: Roll back the last database migration.
- `make migrate_status`: Show current migration status.
- `MIGRATION_NAME=name make migration_sql`: Create a new SQL migration.

## 📄 License

This project is licensed under the [MIT License](LICENSE).
