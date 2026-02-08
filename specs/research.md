# Research: smartLattes

**Date**: 2026-02-08

## R1: Technology Stack Selection

**Decision**: Go (Golang) with embedded static frontend

**Rationale**:
- Produces the smallest Docker image (~25-35MB)
- Single static binary with no runtime dependencies
- `embed` package allows serving static HTML/CSS/JS from within the binary
- Standard library includes production-grade HTTP server and XML parser
- Excellent MongoDB driver (`mongo-go-driver`)
- `net/http` client sufficient for all AI provider REST APIs (no SDKs needed)

**Alternatives considered**:

| Option | Image Size | Rejected Because |
|--------|-----------|------------------|
| Node.js (Fastify) | 150-200MB | Image 4-10x larger than Go |
| Python (FastAPI) | 120-180MB | Image 3-9x larger, slower runtime |
| Bun (Hono) | 90-120MB | Image 3-6x larger, less mature |
| Go | 20-40MB | **Selected** |

## R2: XML Parsing with ISO-8859-1 Encoding

**Decision**: Go `encoding/xml` stdlib + `golang.org/x/text/encoding/charmap`

**Rationale**:
- Go's `encoding/xml` handles XML parsing natively
- `golang.org/x/text` provides ISO-8859-1 (Latin-1) decoder
- Decode ISO-8859-1 to UTF-8 before XML parsing

## R3: Frontend Approach

**Decision**: Vanilla HTML + CSS + minimal JavaScript, embedded in Go binary via `embed`

**Rationale**:
- No build step (no Node.js, no npm)
- Keeps Docker image minimal
- Sufficient for upload form, summary generation form, and results display
- Modern CSS (flexbox/grid, variables)

## R4: MongoDB Driver

**Decision**: `go.mongodb.org/mongo-driver/v2` v2.5.0

**Rationale**: Official driver, upsert support, context-based timeouts, connection pooling.

## R5: HTTP Router

**Decision**: Go standard library `net/http` (Go 1.22+ enhanced routing)

**Rationale**: Method-based routing (`GET /path`, `POST /path`) natively supported. No external dependency for ~8 routes.

## R6: Docker Strategy

**Decision**: Multi-stage build — `golang:1.23-alpine` → `alpine:3.19`

**Rationale**: Static Go binary + alpine base = ~25-35MB image.

## R7: AI Provider Integration — Direct REST API (no SDKs)

**Decision**: Use Go `net/http` to call each provider's REST API directly. No third-party Go SDKs.

**Rationale**:
- All three providers (OpenAI, Anthropic, Gemini) expose simple REST APIs
- Using `net/http` avoids additional dependencies and keeps the Docker image small
- Each provider needs only 2 endpoints: list models + generate completion
- Request/response shapes are simple enough to model with Go structs

**Alternatives considered**:
- Official Go SDKs (openai-go, anthropic-sdk-go, google-generativeai-go): Add dependencies, increase binary size, SDK versions may lag behind API
- Single AI abstraction library (e.g., langchaingo): Over-engineered for 2 API calls

### Provider API Summary

| Aspect | OpenAI | Anthropic | Gemini |
|--------|--------|-----------|--------|
| Base URL | `api.openai.com/v1` | `api.anthropic.com` | `generativelanguage.googleapis.com/v1beta` |
| Auth header | `Authorization: Bearer KEY` | `x-api-key: KEY` + `anthropic-version: 2023-06-01` | `x-goog-api-key: KEY` |
| List models | `GET /models` | `GET /v1/models` | `GET /v1beta/models?key=KEY` |
| Completions | `POST /chat/completions` | `POST /v1/messages` | `POST /v1beta/models/{model}:generateContent` |
| System prompt | `messages[].role="system"` | Top-level `"system"` field | Top-level `"system_instruction"` object |
| Model in request | Body `"model"` field | Body `"model"` field | URL path parameter |
| Response text | `choices[0].message.content` | `content[0].text` | `candidates[0].content.parts[0].text` |
| max_tokens required | No | **Yes** | No (`generationConfig.maxOutputTokens`) |

### Go Implementation Pattern

Each provider implements a common interface:

```go
type AIProvider interface {
    ListModels(ctx context.Context, apiKey string) ([]Model, error)
    GenerateSummary(ctx context.Context, apiKey, model, systemPrompt, userData string) (string, error)
}
```

Three implementations: `openai.go`, `anthropic.go`, `gemini.go` — each using `net/http` with JSON marshal/unmarshal.

## R8: Document Export (Download Formats)

**Decision**: Server-side generation of .md, .docx, and .pdf

**Rationale**:
- **.md**: Trivial — the AI response is already Markdown
- **.docx**: Use `github.com/nguyenthenguyen/docx` or generate manually with Go's `archive/zip` + XML templates (OOXML). Simpler option: use `github.com/unidoc/unioffice` (Apache-2.0 licensed)
- **.pdf**: Use `github.com/jung-kurt/gofpdf` (MIT) for simple PDF generation from text/Markdown

**Alternatives considered**:
- Client-side generation (jsPDF, docx.js): Moves complexity to frontend, harder to control formatting
- Pandoc in Docker: Adds ~300MB to image size, unacceptable

**Selected libraries**:
- PDF: `github.com/jung-kurt/gofpdf` — lightweight, no CGO, MIT license
- DOCX: `github.com/nguyenthenguyen/docx` — simple DOCX generation, MIT license

## R9: CV Search for Dedicated Summary Page

**Decision**: MongoDB text search on `nome-completo` field + regex match on `_id` (lattesID)

**Rationale**:
- Search by lattesID: simple `_id` lookup or prefix regex
- Search by name: MongoDB `$regex` with case-insensitive flag on `curriculo-vitae.dados-gerais.nome-completo`
- For the expected scale (hundreds to low thousands of CVs), regex search is performant enough
- No need for full-text index at this scale

## R10: Token Limit Handling

**Decision**: Estimate token count (~4 chars per token) and truncate CV JSON if it exceeds 80% of model context window

**Rationale**:
- Each provider has different context limits per model
- Use a conservative 4-chars-per-token estimate for truncation threshold
- Truncate least-important sections first: `dados-complementares` → `outra-producao` → `producao-tecnica` → truncate items within `producao-bibliografica`
- Warn user when truncation occurs
