# Tasks: Lattes XML Upload

**Input**: Design documents from `specs/`
**Prerequisites**: plan.md, spec.md, data-model.md, contracts/api.yaml, research.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Go project initialization, dependencies, and directory structure

- [x] T001 Initialize Go module with `go mod init github.com/edalcin/smartlattes` in go.mod
- [x] T002 Add dependencies: `go.mongodb.org/mongo-driver/v2` and `golang.org/x/text` in go.mod
- [x] T003 Create directory structure: `cmd/smartlattes/`, `internal/handler/`, `internal/parser/`, `internal/store/`, `internal/static/css/`, `internal/static/js/`
- [x] T004 [P] Create .env.example with generic MONGODB_URI, MONGODB_DATABASE, PORT variables in .env.example
- [x] T005 [P] Create Dockerfile with multi-stage build (golang:1.23-alpine → alpine:3.19) in Dockerfile

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T006 Implement MongoDB client: connect (with MONGODB_URI env var, fail if not set), ping, disconnect, get collection in internal/store/mongo.go
- [x] T007 Implement Lattes XML parser: ISO-8859-1 decoding via `golang.org/x/text/encoding/charmap`, recursive XML-to-map converter (attributes with `@` prefix, repeated elements as arrays), validation (check CURRICULO-VITAE root element and NUMERO-IDENTIFICADOR attribute) in internal/parser/lattes.go
- [x] T008 Implement environment config loader: read MONGODB_URI (required), MONGODB_DATABASE (default "smartLattes"), PORT (default "8080"), MAX_UPLOAD_SIZE (default 10485760) in cmd/smartlattes/main.go
- [x] T009 Create shared CSS stylesheet: CSS variables for colors/spacing, flexbox layout, responsive design, modern typography, form styling, navigation bar styling, message/alert components in internal/static/css/style.css

**Checkpoint**: Foundation ready — MongoDB client connects, XML parser converts Lattes XML to map, env config loads, CSS ready

---

## Phase 3: User Story 1 — Upload Lattes XML File (Priority: P1) MVP

**Goal**: A researcher uploads a Lattes XML file, the system parses it, converts to JSON, stores in MongoDB, and shows a success message with name and Lattes ID.

**Independent Test**: Upload `docs/8334174268306003.xml` via the web form and verify the document appears in MongoDB collection `curriculos` with `_id: "8334174268306003"`.

### Implementation for User Story 1

- [x] T010 [US1] Implement MongoDB upsert operation: ReplaceOne with upsert=true, using NUMERO-IDENTIFICADOR as `_id`, include `_metadata` field (uploadedAt, originalFilename, fileSize) in internal/store/mongo.go
- [x] T011 [US1] Create upload HTML page: file input accepting .xml files, submit button, drag-and-drop zone, inline result/error display area in internal/static/upload.html
- [x] T012 [US1] Implement upload JavaScript: FormData submission via fetch to POST /api/upload, display success (name, Lattes ID) or error message inline, loading state during upload in internal/static/js/upload.js
- [x] T013 [US1] Implement upload HTTP handler: accept multipart file, validate size (max 10MB → 413), validate not empty (→ 400), call parser to validate and convert XML, call store to upsert document, return UploadSuccess or Error JSON per contracts/api.yaml in internal/handler/upload.go
- [x] T014 [US1] Wire up main.go: load env config, connect to MongoDB, register routes (POST /api/upload → upload handler, serve static files via embed), graceful shutdown on SIGTERM in cmd/smartlattes/main.go

**Checkpoint**: Upload works end-to-end. Researcher can upload XML, see success with name and ID, data stored in MongoDB.

---

## Phase 4: User Story 2 — View Upload Confirmation Details (Priority: P2)

**Goal**: After successful upload, show a summary with full name, Lattes ID, last update date, and counts of production sections (bibliographic, technical, other).

**Independent Test**: Upload `docs/8334174268306003.xml` and verify the confirmation shows name "Eduardo Couto Dalcin", ID "8334174268306003", update date "03012026", and correct counts for each production section.

### Implementation for User Story 2

- [x] T015 [US2] Extend parser to extract summary data: full name (DADOS-GERAIS/@NOME-COMPLETO), Lattes ID, DATA-ATUALIZACAO, count children in PRODUCAO-BIBLIOGRAFICA, PRODUCAO-TECNICA, OUTRA-PRODUCAO in internal/parser/lattes.go
- [x] T016 [US2] Extend upload handler response to include counts object (bibliographicProduction, technicalProduction, otherProduction) per UploadSuccess schema in internal/handler/upload.go
- [x] T017 [US2] Update upload JavaScript to display detailed confirmation: name, Lattes ID, formatted update date, production section counts in a summary card in internal/static/js/upload.js

**Checkpoint**: Upload shows detailed confirmation with name, ID, date, and production counts.

---

## Phase 5: User Story 3 — Navigate Application via Main Menu (Priority: P3)

**Goal**: Homepage with navigation menu linking to Upload CV (active) and Data Explorer (coming soon placeholder).

**Independent Test**: Navigate to homepage, see menu with two items, click "Enviar CV" → upload page loads, click "Data Explorer" → placeholder page loads.

### Implementation for User Story 3

- [x] T018 [P] [US3] Create homepage HTML: navigation bar with "smartLattes" branding, menu items "Enviar CV" (link to /upload) and "Explorar Dados" (link to /explorer), hero section with brief description of the application in internal/static/index.html
- [x] T019 [P] [US3] Create Data Explorer placeholder page: same navigation bar, "coming soon" message with icon, brief description of future functionality in internal/static/explorer.html
- [x] T020 [US3] Implement page handler: serve index.html on GET /, upload.html on GET /upload, explorer.html on GET /explorer using Go embed and http.FileServer in internal/handler/pages.go
- [x] T021 [US3] Register page routes in main.go: GET / → pages handler, GET /upload → pages handler, GET /explorer → pages handler, static assets (css/, js/) in cmd/smartlattes/main.go

**Checkpoint**: All navigation works. Homepage shows menu, Upload and Explorer pages are accessible.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Health check, Docker validation, final integration

- [x] T022 [P] Implement health check handler: GET /api/health returns {"status":"healthy","mongodb":"connected"} or {"status":"unhealthy","mongodb":"disconnected"} with appropriate HTTP status in internal/handler/health.go
- [x] T023 [P] Register health route in main.go: GET /api/health → health handler in cmd/smartlattes/main.go
- [x] T024 Build Docker image and verify: `docker build -t ghcr.io/edalcin/smartlattes:latest .`, verify image size < 200MB, verify container starts and responds on port 8080 in Dockerfile
- [x] T025 End-to-end validation: start container with MONGODB_URI env var, upload docs/8334174268306003.xml via browser, verify MongoDB document, verify confirmation details, test navigation, test error cases (invalid file, oversized file, empty file)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Setup completion — BLOCKS all user stories
- **US1 (Phase 3)**: Depends on Foundational — core upload functionality
- **US2 (Phase 4)**: Depends on US1 (extends parser and handler from US1)
- **US3 (Phase 5)**: Depends on Foundational only — can run in parallel with US1/US2
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 only. No dependencies on other stories.
- **US2 (P2)**: Depends on US1 (extends the parser and upload handler created in US1).
- **US3 (P3)**: Depends on Phase 2 only. Can be implemented in parallel with US1.

### Within Each User Story

- Store operations before handlers (handler calls store)
- Parser before handler (handler calls parser)
- HTML/JS before or in parallel with handler (different files)
- Handler wiring in main.go is last step

### Parallel Opportunities

- T004 and T005 can run in parallel (Setup phase)
- T018 and T019 can run in parallel (US3 — different HTML files)
- T022 and T023 can run in parallel with US3 tasks (different files)
- US1 and US3 can be developed in parallel after Phase 2

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001–T005)
2. Complete Phase 2: Foundational (T006–T009)
3. Complete Phase 3: User Story 1 (T010–T014)
4. **STOP and VALIDATE**: Upload XML via minimal interface, verify MongoDB storage
5. Deploy if ready — basic upload works

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add US1 → Upload works → **MVP deployed**
3. Add US2 → Confirmation details shown → Deploy update
4. Add US3 → Full navigation → Deploy update
5. Add Polish → Health check, Docker image optimized → Final release

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- MONGODB_URI is required — app must fail to start without it
- The `.env` file with real credentials is in `.gitignore` — never commit it
- Use `docs/8334174268306003.xml` as the primary test artifact for all stories
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
