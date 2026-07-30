Here is a complete, fresh `README.md` tailored specifically to the Go backend service code and task configurations provided:

# Media Diary - Go Backend Service

A high-performance Go HTTP REST service for managing media diary entries, calculating title ratings and access statistics, and processing batch CSV data imports/exports with PostgreSQL.

---

## 🛠 Features

* **Media Entry CRUD**: Full endpoints to fetch, add, edit, and delete entries.
* **Analytics & Rating**: Title statistics across custom timeframes and aggregated media ratings.
* **Bulk Import & Export**:
  * Import entries from semicolon-separated (`;`) CSV files inside transactional database operations.
  * Export computed rating summaries to local CSV files (`export/`).
* **Search Capabilities**: Query entries matching title, type, or comment patterns (`ILIKE`).
* **Database Migrations**: Integrated database migration tasks using [`goose`](https://github.com/pressly/goose) and [`mise`](https://mise.jdx.dev/).

---

## 🏗 Architecture & Project Structure

The project follows a standard layered architecture:

```
backend-go/
├── config/         # Environment variable loading & app configurations
├── db/             # PostgreSQL database connection setup
├── handler/        # HTTP handlers & REST route definitions (net/http ServeMux)
├── model/          # Data models, structs, & custom JSON/DB drivers (LocalDate)
├── repository/     # Raw SQL query execution & database interface implementations
├── service/        # Core business logic, CSV parsing, and scoring algorithms
└── mise.toml       # Environment configuration & task runner commands
```

---

## ⚙️ Environment Variables

The service loads configurations from environment variables with sensible defaults:

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT_GO` | Port on which the HTTP server listens | `8080` |
| `PORT_DB` | PostgreSQL database port | `5432` |
| `DB_HOST` | Database host address | `localhost` |
| `DB_USER` | Database username | `postgres` |
| `DB_PASSWORD` | Database password | `postgres` |
| `DB_NAME` | Database name | `matalogue` |
| `ALLOWED_ORIGIN` | CORS allowed origin header | `http://localhost:3000` |

---

## 🗄 Database & Migrations

Database migrations are managed with **Goose** and executed via **Mise**.

### Migration Commands

Make sure you have [mise](https://mise.jdx.dev/) installed on your machine.

* **Check status:**
```bash
    mise run migrate:status
```

* **Apply all pending migrations:**
```bash
    mise run migrate:up
```

* **Roll back last migration:**
```bash
    mise run migrate:down
```

* **Create a new migration file:**
```bash
    mise run migrate:create -- <migration_name> sql
```

---

## 📡 API Endpoints

### Media Entries

* `GET /api/entries/all` — Fetch all media entries ordered by date descending.
* `GET /api/entries/date/{date}` — Fetch entries for a specific date (`YYYY-MM-DD`).
* `GET /api/entries/{days}` — Fetch entries recorded on or after a target date.
* `POST /api/entries` — Add a new media entry.
* `PUT /api/entries/{entryId}` — Update an existing media entry.
* `DELETE /api/entries/{entryId}` — Delete a media entry by ID.

### Analytics & Search

* `GET /api/stats` — Calculate detailed lookup statistics for a specific title.
* `GET /api/entries/ratings/{months}` — Get aggregated title ratings filtered by a month range.
* `GET /api/entries/ratings/{months}/export` — Export calculated ratings to a local CSV file.
* `GET /api/entries/search/{searchTerm}` — Search entries across title, comment, or media type.

### Bulk Operations

* `POST /api/entries/import` — Import and replace active entries using a semicolon-delimited CSV payload (`DD.MM.YYYY` date format).

---

## 🚀 Getting Started

### Prerequisites

* **Go** 1.22 or higher
* **PostgreSQL** instance running locally or via Docker

### Running Locally

1. **Install dependencies:**
```bash
    go mod tidy
```

2. **Apply database migrations:**
```bash
    mise run migrate:up
```

3. **Start the application:**
```bash
    go run main.go
```
