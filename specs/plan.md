# Implementation Plan: smartLattes

**Branch**: `main` | **Date**: 2026-02-08 | **Spec**: [spec.md](spec.md)

## Summary

Extend the existing smartLattes web application (Go + MongoDB + vanilla HTML/CSS/JS) with the **Transformation context**: AI-powered researcher summary generation. The system integrates with three AI providers (OpenAI, Anthropic, Google Gemini) via direct REST API calls, presents a model selection UI, generates researcher summaries, and stores them in MongoDB. The prompt is maintained in an editable `resumoPrompt.md` file. Users can download summaries in .md, .docx, or .pdf formats.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: `go.mongodb.org/mongo-driver/v2`, `golang.org/x/text`, `github.com/jung-kurt/gofpdf`, `github.com/nguyenthenguyen/docx`
**Storage**: MongoDB (external, database `smartLattes`, collections: `curriculos`, `resumos`)
**AI Integration**: Direct REST API calls via `net/http` — no third-party SDKs
**Testing**: `go test` (stdlib)
**Target Platform**: Linux container (Docker, `ghcr.io/edalcin/`)
**Project Type**: Single project (Go backend with embedded static frontend)
**Performance Goals**: Upload within 10s; AI generation within 120s timeout; container start < 10s
**Constraints**: Docker image < 200MB; single container; no authentication; HTTP only; API keys not persisted
**Scale/Scope**: Internal/institutional use, low concurrent users

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file contains only the blank template — no project-specific principles defined. **GATE PASSED** (no constraints to violate).

**Post-Phase 1 re-check**: Design adds 2 new internal packages (`ai`, `export`) to existing 3 (`handler`, `parser`, `store`). AI provider calls are isolated behind an interface. No violations.

## Project Structure

### Documentation

```text
specs/
├── spec.md              # Feature specification (updated)
├── plan.md              # This file (updated)
├── research.md          # Technology research (updated)
├── data-model.md        # MongoDB document model (updated)
├── quickstart.md        # Development setup guide (updated)
├── contracts/
│   └── api.yaml         # OpenAPI 3.1 contract (updated)
├── checklists/
│   └── requirements-checklist.md
└── tasks.md             # Task list (/speckit.tasks output)
```

### Source Code (repository root)

```text
cmd/
└── smartlattes/
    └── main.go              # Entry point: server setup, routing, graceful shutdown

internal/
├── handler/
│   ├── upload.go            # POST /api/upload (existing)
│   ├── pages.go             # GET /, /upload, /resumo, /explorer (updated)
│   ├── health.go            # GET /api/health (existing)
│   ├── models.go            # POST /api/models — list AI models for provider+key
│   ├── summary.go           # POST /api/summary — generate researcher summary
│   ├── search.go            # GET /api/search?q= — search CVs by name/lattesID
│   └── download.go          # GET /api/download/{lattesID}?format= — download summary
├── parser/
│   └── lattes.go            # XML parsing (existing, updated: lowercase + filter)
├── store/
│   └── mongo.go             # MongoDB operations (updated: resumos collection, search)
├── ai/
│   ├── provider.go          # AIProvider interface + Model type + factory function
│   ├── openai.go            # OpenAI REST API implementation
│   ├── anthropic.go         # Anthropic REST API implementation
│   ├── gemini.go            # Google Gemini REST API implementation
│   └── truncate.go          # Token estimation + CV data truncation logic
└── export/
    ├── markdown.go          # .md export (trivial — raw text)
    ├── pdf.go               # .pdf generation via gofpdf
    └── docx.go              # .docx generation

internal/static/
├── index.html               # Homepage with 3 menu items (updated)
├── upload.html              # Upload form (updated: add summary form after success)
├── resumo.html              # NEW: Dedicated summary generation page (search + form)
├── explorer.html            # "Coming soon" placeholder (existing)
├── css/
│   └── style.css            # Styles (updated: summary form, results display)
└── js/
    ├── upload.js            # Upload form behavior (updated: summary flow after upload)
    └── resumo.js            # NEW: Summary page — search, provider/model selection, generation, download

resumoPrompt.md              # AI prompt for summary generation (NEW, repo root)
```

**Structure Decision**: Continue single Go project pattern. Add 2 new internal packages:
- `ai/` — Provider abstraction with interface + 3 implementations + truncation
- `export/` — Document format conversion (.md, .docx, .pdf)

New handlers added to `handler/` package (one file per endpoint). New frontend page `resumo.html` + `resumo.js` for the dedicated summary page.

## Key Design Decisions

### 1. XML-to-JSON Conversion (existing, updated)

- Generic recursive XML-to-map converter
- All keys stored in **lowercase** (new)
- `dados-gerais` filtered to only 6 allowed fields (new)
- No `@` prefix on attributes (changed from original plan)

### 2. AI Provider Abstraction

Common interface isolates provider-specific logic:

```go
type AIProvider interface {
    ListModels(ctx context.Context, apiKey string) ([]Model, error)
    Generate(ctx context.Context, req GenerateRequest) (string, error)
}

type Model struct {
    ID          string `json:"id"`
    DisplayName string `json:"displayName"`
}

type GenerateRequest struct {
    APIKey       string
    Model        string
    SystemPrompt string
    UserData     string
    MaxTokens    int
}
```

Factory function selects implementation by provider name:

```go
func NewProvider(name string) (AIProvider, error)
// name: "openai" | "anthropic" | "gemini"
```

### 3. REST API Calls (no SDKs)

Each provider implementation uses `net/http` directly:
- **OpenAI**: `Authorization: Bearer KEY`, `POST /v1/chat/completions`
- **Anthropic**: `x-api-key: KEY` + `anthropic-version: 2023-06-01`, `POST /v1/messages`
- **Gemini**: `x-goog-api-key: KEY`, `POST /v1beta/models/{model}:generateContent`

All calls use a shared 120-second timeout context.

### 4. Model Listing Flow

1. User selects provider in pulldown
2. User types API key
3. Frontend calls `POST /api/models` with `{provider, apiKey}`
4. Backend calls provider's list-models endpoint
5. Backend returns filtered model list (only chat/completion capable models)
6. Frontend populates second pulldown with models

### 5. Summary Generation Flow

1. User has lattesID (from upload or search)
2. User selects provider + model + enters API key
3. Frontend calls `POST /api/summary` with `{lattesID, provider, apiKey, model}`
4. Backend:
   a. Reads CV from `curriculos` collection
   b. Reads `resumoPrompt.md` from embedded filesystem
   c. Serializes CV data to JSON string
   d. Estimates tokens; truncates if needed (FR-110)
   e. Calls AI provider with system=prompt, user=CV data
   f. Returns generated summary text + truncation warning if applicable
5. Frontend displays summary with download buttons
6. On "Ok" or download: frontend calls `POST /api/summary/save` to persist to `resumos` collection

### 6. Token Truncation Strategy (FR-110)

Estimate: ~4 chars per token. Truncation order (least to most important):
1. Remove `dados-complementares`
2. Remove `outra-producao`
3. Remove `producao-tecnica`
4. Truncate items within `producao-bibliografica` (keep most recent)

Response includes `truncated: true` flag + warning message when truncation occurs.

### 7. Document Export

All export happens server-side:
- **GET /api/download/{lattesID}?format=md** — returns raw Markdown as `text/markdown`
- **GET /api/download/{lattesID}?format=pdf** — generates PDF via gofpdf
- **GET /api/download/{lattesID}?format=docx** — generates DOCX

Export reads the summary from the `resumos` MongoDB collection.

### 8. Frontend Approach

Two entry points for summary generation:
1. **After upload** (`upload.html`): Success screen shows summary form inline
2. **Dedicated page** (`resumo.html`): Search field → find CV → show summary form

Both use the same API endpoints. The `resumo.js` handles:
- CV search with debounced API calls
- Provider selection → API key input → model listing
- Summary generation with loading state
- Summary display with Markdown rendering
- Download buttons (3 formats)

### 9. MongoDB Connection (existing)

`MONGODB_URI` via environment variable. Connection on startup with 5-second timeout.

### 10. Error Handling Strategy (updated)

All API errors return JSON `{ "success": false, "error": "message" }`:
- **400**: Invalid request, missing fields
- **401**: Invalid API key (from AI provider)
- **413**: File exceeds 10MB
- **404**: CV not found (for summary generation)
- **503**: MongoDB unavailable
- **504**: AI provider timeout (120s)

### 11. Prompt Management

`resumoPrompt.md` is embedded at compile time via `//go:embed`. To update the prompt:
1. Edit `resumoPrompt.md` in repo root
2. Rebuild Docker image
3. Redeploy

This keeps the prompt versioned with the code while being easy to edit.

## Complexity Tracking

No constitution violations — no entries needed.
