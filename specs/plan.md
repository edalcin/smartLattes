# Implementation Plan: Lattes XML Upload

**Branch**: `main` | **Date**: 2026-02-07 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/spec.md`

## Summary

Build a web application that allows researchers to upload their Lattes CV XML files (exported from CNPq's Plataforma Lattes), converts them to JSON, and stores them in MongoDB. The application uses Go for a minimal Docker image (~25-35MB), serves static HTML/CSS/JS frontend embedded in the binary, and provides a clean, modern upload interface with navigation menu.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: `go.mongodb.org/mongo-driver/v2`, `golang.org/x/text` (ISO-8859-1 encoding)
**Storage**: MongoDB 6+ (external, at `<host>:27017`, database `smartLattes`)
**Testing**: `go test` (stdlib)
**Target Platform**: Linux container (Docker, published to `ghcr.io/edalcin/`)
**Project Type**: Single project (Go backend with embedded static frontend)
**Performance Goals**: Upload + parse + store within 10 seconds for files up to 1MB; container start < 10 seconds
**Constraints**: Docker image < 200MB (target ~25-35MB); single container; no authentication; deployed on Unraid via web UI
**Scale/Scope**: Internal/institutional use, low concurrent users, single MongoDB collection

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file contains only the blank template — no project-specific principles defined yet. **GATE PASSED** (no constraints to violate).

**Post-Phase 1 re-check**: Design uses a single Go project with 3 internal packages (`handler`, `parser`, `store`) — minimal complexity. No violations.

## Project Structure

### Documentation

```text
specs/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0: technology research
├── data-model.md        # Phase 1: MongoDB document model
├── quickstart.md        # Phase 1: development setup guide
├── contracts/
│   └── api.yaml         # Phase 1: OpenAPI 3.1 contract
├── checklists/
│   └── requirements-checklist.md
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
cmd/
└── smartlattes/
    └── main.go              # Entry point: server setup, routing, graceful shutdown

internal/
├── handler/
│   ├── upload.go            # POST /api/upload — file validation, parse, store, response
│   ├── pages.go             # GET /, /upload, /explorer — serve HTML pages
│   └── health.go            # GET /api/health — MongoDB connectivity check
├── parser/
│   └── lattes.go            # XML parsing: ISO-8859-1 decode, XML-to-map conversion, validation
└── store/
    └── mongo.go             # MongoDB client: connect, upsert CV document, health ping

internal/static/              # Embedded via Go embed directive
├── index.html               # Homepage with navigation menu
├── upload.html              # Upload form page (result displayed inline via JS)
├── explorer.html            # "Coming soon" placeholder
├── css/
│   └── style.css            # Modern CSS (variables, flexbox, responsive)
└── js/
    └── upload.js            # Upload form: file selection, submit, display result

docs/
└── 8334174268306003.xml     # Example Lattes XML for testing

Dockerfile                   # Multi-stage: golang:1.23-alpine → alpine:3.19
go.mod
go.sum
```

**Structure Decision**: Single Go project. The backend serves both the API (`/api/*`) and the static frontend files. No separate frontend build. The `internal/` directory follows Go conventions with three focused packages: `handler` (HTTP), `parser` (XML logic), `store` (MongoDB). Static files are embedded at compile time using `//go:embed`.

## Complexity Tracking

No constitution violations — no entries needed.

## Key Design Decisions

### 1. XML-to-JSON Conversion

Use a recursive generic XML-to-map converter (not struct-based) because:
- Lattes XML has hundreds of distinct elements — defining Go structs for all would be impractical
- A generic converter preserves the complete XML structure automatically
- Attributes are prefixed with `@` (e.g., `@NOME-COMPLETO`)
- Repeated child elements become JSON arrays
- See `data-model.md` for the output structure

### 2. Frontend Approach

Vanilla HTML/CSS/JS with Go `html/template` for dynamic content:
- Upload form uses `fetch()` with `FormData` for async upload
- Result page rendered server-side after processing
- Modern CSS with Pico CSS-inspired styling (custom, no external dependency)
- Responsive design for desktop and mobile
- No JavaScript framework needed — the UI has 4 pages total

### 3. Docker Multi-Stage Build

```dockerfile
# Stage 1: Build
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o smartlattes ./cmd/smartlattes

# Stage 2: Runtime
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /build/smartlattes /smartlattes
EXPOSE 8080
CMD ["/smartlattes"]
```

Expected image size: ~25-35MB.

### 4. Error Handling Strategy

All errors return JSON with `{ "success": false, "error": "message" }`:
- **400**: Invalid file (not XML, not Lattes, empty, wrong structure)
- **413**: File exceeds 10MB
- **503**: MongoDB unavailable
- Frontend displays error messages inline on the upload page

### 5. MongoDB Connection

- Connection string provided exclusively via `MONGODB_URI` environment variable (no default — app MUST exit immediately if env var is not set)
- Connect on startup with 5-second timeout
- If env var is set but MongoDB is temporarily unreachable on startup: log warning, continue running (health check reports status; uploads return 503 until connection is established)
- If connection fails during upload: return 503 immediately
- In production, `MONGODB_URI` is configured via Unraid Docker UI as a container environment variable

### 6. Deployment Platform: Unraid

- The Docker container runs on Unraid, configured via its web interface
- No docker-compose or orchestration needed — Unraid manages the container directly
- Port mapping, environment variables, and updates are all configured through the Unraid Docker UI
- See `quickstart.md` for step-by-step Unraid deployment instructions
