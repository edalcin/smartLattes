# Implementation Plan: smartLattes — Contexto de Análise

**Branch**: `main` | **Date**: 2026-02-09 | **Spec**: [spec.md](../spec.md)
**Input**: Feature specification from `/specs/spec.md` (User Stories 6, 7; FR-200 to FR-209)

## Summary

Extend the smartLattes web application with the **Analysis context**: AI-powered researcher relationship identification. After generating a summary, the system offers to analyze the researcher's CV against all other researchers in MongoDB, using the same AI provider/model/key. The AI generates two lists — researchers with common interests and researchers with complementary interests (with interdisciplinary project suggestions). Results are displayed, downloadable (.md/.docx/.pdf), and stored in the `relacoes` collection. The analysis prompt is maintained in an editable `analisePrompt.md` file, embedded at compile time.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: `go.mongodb.org/mongo-driver/v2`, `golang.org/x/text`, `net/http` (AI REST calls)
**Storage**: MongoDB (external, database `smartLattes`, collections: `curriculos`, `resumos`, `relacoes`)
**AI Integration**: Direct REST API calls via `net/http` — same providers (OpenAI, Anthropic, Gemini)
**Testing**: `go test` (stdlib), Docker build validation
**Target Platform**: Linux container (Docker, `ghcr.io/edalcin/`)
**Project Type**: Single project (Go backend with embedded static frontend)
**Performance Goals**: Analysis generation within 120s timeout; container start < 10s
**Constraints**: Docker image < 200MB; single container; no auth; HTTP only; API keys not persisted; reuse provider/model/key from summary step
**Scale/Scope**: Internal/institutional use, low concurrent users; analysis compares 1 researcher against N others in the base

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Constitution file contains only the blank template — no project-specific principles defined. **GATE PASSED** (no constraints to violate).

**Post-Phase 1 re-check**: Design extends existing packages (`handler`, `store`, `ai`, `export`) and adds one embedded file (`analisePrompt.md`). No new packages. No violations.

## Project Structure

### Documentation

```text
specs/
├── spec.md              # Feature specification (updated with Análise context)
├── plan.md              # Original plan (Transformação)
├── data-model.md        # MongoDB document model (to be updated with relacoes)
├── research.md          # Technology research (existing)
├── quickstart.md        # Development setup guide (existing)
├── contracts/
│   └── api.yaml         # OpenAPI 3.1 contract (to be updated)
└── tasks.md             # Task list (to be generated via /speckit.tasks)

specs/main/
├── plan.md              # This file (Análise context plan)
├── research.md          # Phase 0 research output
├── data-model.md        # Phase 1 data model (relacoes collection)
├── quickstart.md        # Phase 1 quickstart (analysis-specific)
└── contracts/
    └── api.yaml         # OpenAPI contract for analysis endpoints
```

### Source Code (repository root)

```text
cmd/
└── smartlattes/
    ├── main.go              # Entry point: updated routes for analysis endpoints
    ├── resumoPrompt.md      # Existing: summary prompt
    └── analisePrompt.md     # NEW: analysis prompt (embedded at compile time)

internal/
├── handler/
│   ├── upload.go            # Existing (no changes)
│   ├── pages.go             # Updated: serve analise.html page
│   ├── health.go            # Existing (no changes)
│   ├── models.go            # Existing (no changes)
│   ├── summary.go           # Existing (no changes)
│   ├── search.go            # Existing (no changes)
│   ├── download.go          # Updated: handle analysis report download from relacoes
│   ├── analysis.go          # NEW: POST /api/analysis — generate relationship analysis
│   │                        #       POST /api/analysis/save — persist to MongoDB
│   └── helpers.go           # Existing (no changes)
├── parser/
│   └── lattes.go            # Existing (no changes)
├── store/
│   └── mongo.go             # Updated: relacoes collection methods
├── ai/
│   ├── provider.go          # Existing (no changes to interface)
│   ├── openai.go            # Existing (no changes)
│   ├── anthropic.go         # Existing (no changes)
│   ├── gemini.go            # Existing (no changes)
│   └── truncate.go          # Updated: add TruncateMultipleCVs for analysis context
└── export/
    └── markdown.go          # Existing (reused for analysis report)

internal/static/
├── index.html               # Updated: add "Analisar Relações" menu item
├── upload.html              # Updated: add analysis prompt after summary generation
├── resumo.html              # Updated: add analysis prompt after summary generation
├── analise.html             # NEW: dedicated analysis page
├── explorer.html            # Existing placeholder
├── css/
│   └── style.css            # Updated: analysis-specific styles
└── js/
    ├── upload.js            # Updated: add analysis flow after summary
    ├── resumo.js            # Updated: add analysis flow after summary
    └── analise.js           # NEW: dedicated analysis page logic
```

**Structure Decision**: Continue single Go project pattern. No new internal packages — extend existing `handler`, `store`, `ai`, and `export`. Add one new handler file (`analysis.go`), one new HTML page (`analise.html`), one new JS file (`analise.js`), and one new embedded prompt (`analisePrompt.md`).

## Key Design Decisions

### 1. Analysis Flow (FR-200 to FR-209)

After the summary is generated (in both `upload.html` and `resumo.html`):

1. UI shows: "Deseja analisar relações com outros pesquisadores?" [Sim] [Não]
2. If "Sim":
   a. Frontend calls `POST /api/analysis` with `{lattesId, provider, apiKey, model}`
   b. Backend reads current researcher's CV from `curriculos`
   c. Backend reads ALL other CVs from `curriculos` (excluding current)
   d. If no other CVs exist → return error "Não há outros pesquisadores na base" (FR-206)
   e. Backend truncates combined data if needed (FR-208)
   f. Backend calls AI provider with system=`analisePrompt.md`, user=combined CVs JSON
   g. Returns analysis text + truncation warning if applicable
3. Frontend displays two lists (rendered from Markdown)
4. Download buttons (.md, .docx, .pdf) via `GET /api/analysis/download/{lattesId}?format=`
5. On download or "Ok": frontend calls `POST /api/analysis/save` to persist to `relacoes`
6. If "Não": flow ends, no analysis generated

### 2. Data Retrieval Strategy

```go
// New method in store/mongo.go
GetAllCVSummaries(ctx, excludeLattesID string) ([]CVAnalysisData, error)
```

This method retrieves all CVs except the current researcher's, projecting only the fields relevant for analysis:
- `_id` (lattesID)
- `curriculo-vitae.dados-gerais.nome-completo`
- `curriculo-vitae.dados-gerais.areas-de-atuacao`
- `curriculo-vitae.dados-gerais.formacao-academica-titulacao`
- `curriculo-vitae.dados-gerais.atuacoes-profissionais`
- `curriculo-vitae.producao-bibliografica` (titles/topics only, for matching)

Projection minimizes data transfer and token usage.

### 3. Token Management for Multi-CV Analysis

```go
// New function in ai/truncate.go
TruncateAnalysisData(currentCV, otherCVs []map[string]interface{}, maxTokens int) (string, bool)
```

Strategy:
1. Always include full current researcher data
2. For other CVs: include essential fields (name, areas, formation)
3. If still over limit: reduce production data progressively
4. If still over limit: limit number of researchers compared (most relevant first)
5. Return combined JSON + truncation flag

### 4. Analysis Prompt Management

`analisePrompt.md` is placed in `cmd/smartlattes/` (same as `resumoPrompt.md`) for `//go:embed` compatibility. Embedded at compile time:

```go
//go:embed analisePrompt.md
var analisePrompt string
```

The prompt instructs the AI to:
- Analyze the target researcher against all others provided
- Generate two Markdown sections:
  - "Pesquisadores com Interesses Comuns" — name, lattesID, common areas, shared production themes
  - "Pesquisadores com Interesses Complementares" — name, lattesID, complementary areas, suggested interdisciplinary projects
- Output in Portuguese (BR)
- Be data-driven, cite specific areas/publications

### 5. MongoDB `relacoes` Collection

```json
{
  "_id": "8334174268306003",
  "analise": "# Análise de Relações\n\n## Pesquisadores com Interesses Comuns\n...\n\n## Pesquisadores com Interesses Complementares\n...",
  "_metadata": {
    "generatedAt": "2026-02-09T14:00:00Z",
    "provider": "anthropic",
    "model": "claude-sonnet-4-5-20250929",
    "researchersAnalyzed": 15
  }
}
```

New store methods:
```go
UpsertAnalysis(ctx, lattesID, analysis string, provider, model string, count int) error
GetAnalysis(ctx, lattesID) (*AnalysisDoc, error)
CountCVs(ctx) (int64, error)  // For checking if > 1 researcher exists
```

### 6. API Endpoints (New)

| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/analysis` | Generate relationship analysis |
| POST | `/api/analysis/save` | Persist analysis to MongoDB |
| GET | `/api/analysis/download/{lattesId}?format=` | Download analysis report |

**POST /api/analysis** request:
```json
{
  "lattesId": "8334174268306003",
  "provider": "anthropic",
  "apiKey": "sk-...",
  "model": "claude-sonnet-4-5-20250929"
}
```

**POST /api/analysis** response:
```json
{
  "success": true,
  "analysis": "# Análise de Relações\n...",
  "provider": "anthropic",
  "model": "claude-sonnet-4-5-20250929",
  "researchersAnalyzed": 15,
  "truncated": false,
  "truncationWarning": ""
}
```

### 7. Frontend — Analysis Prompt After Summary

In both `upload.js` and `resumo.js`, after summary is displayed:

1. Show a card: "Deseja analisar relações com outros pesquisadores?"
2. Two buttons: "Sim, analisar" / "Não, obrigado"
3. If "Sim": show progress indicator, call `/api/analysis`, display results
4. Analysis results rendered with same Markdown renderer
5. Download buttons + "Ok" button (triggers save)
6. Reuses `currentLattesId`, `currentProvider`, `currentApiKey`, `currentModel` from summary step

### 8. Dedicated Analysis Page (`analise.html`)

For running analysis independently (from menu "Analisar Relações"):
- Search field to find researcher (same as resumo.html)
- Provider/key/model selection (same pattern)
- Generate analysis button
- Results display with download

### 9. Download Handler Extension

Extend existing `DownloadHandler` or create parallel `AnalysisDownloadHandler`:
- `GET /api/analysis/download/{lattesId}?format=md|docx|pdf`
- Reads from `relacoes` collection instead of `resumos`
- Same export logic (.md implemented, .docx/.pdf stub 501)

### 10. Error Handling (FR-207, FR-208, FR-209)

| Error | HTTP | Message |
|-------|------|---------|
| No other researchers in base | 409 | "Não há outros pesquisadores na base para comparação" |
| Invalid API key | 401 | (from AI provider) |
| Provider timeout | 504 | "Tempo limite excedido na análise" |
| Provider unavailable | 503 | (from AI provider) |
| MongoDB unavailable | 503 | "Banco de dados indisponível" |
| Data truncated | 200 | Success with `truncated: true` + warning |

### 11. Progress Indicator (FR-209)

Frontend shows animated spinner with text "Analisando relações entre pesquisadores..." during the `/api/analysis` call. Same pattern as existing summary generation spinner.

## Complexity Tracking

No constitution violations — no entries needed.
