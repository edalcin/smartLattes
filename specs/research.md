# Research: Lattes XML Upload

**Date**: 2026-02-07

## R1: Technology Stack Selection

**Decision**: Go (Golang) with embedded static frontend

**Rationale**:
- Produces the smallest Docker image by far (~20-40MB vs 120-200MB for alternatives)
- Single static binary with no runtime dependencies
- `embed` package allows serving static HTML/CSS/JS from within the binary
- Standard library includes production-grade HTTP server and XML parser
- Excellent MongoDB driver (`mongo-go-driver`)
- Meets the "simple but modern" requirement — Go is a modern language with simple syntax
- Can run on `scratch` or `alpine` base image

**Alternatives considered**:

| Option | Image Size | Simplicity | Rejected Because |
|--------|-----------|------------|------------------|
| Node.js (Fastify) | 150-200MB | 5/5 | Image 4-10x larger than Go |
| Python (FastAPI) | 120-180MB | 5/5 | Image 3-9x larger, slower runtime |
| Bun (Hono) | 90-120MB | 4/5 | Image 3-6x larger, less mature ecosystem |
| Go | 20-40MB | 3/5 | **Selected** |

## R2: XML Parsing with ISO-8859-1 Encoding

**Decision**: Use Go `encoding/xml` stdlib + `golang.org/x/text/encoding/charmap` for ISO-8859-1 decoding

**Rationale**:
- Go's `encoding/xml` handles XML parsing natively
- The `golang.org/x/text` package provides ISO-8859-1 (Latin-1) decoder
- Decode ISO-8859-1 to UTF-8 before XML parsing
- Lattes XML uses `encoding="ISO-8859-1"` declaration — Go respects this with a custom `CharsetReader`

**Alternatives considered**:
- `etree` library: More complex API than needed for simple attribute/element extraction
- Manual byte conversion: Error-prone, not standards-compliant

## R3: Frontend Approach

**Decision**: Vanilla HTML + CSS + minimal JavaScript, embedded in Go binary via `embed` package

**Rationale**:
- No build step required (no Node.js, no npm, no bundler)
- Keeps Docker image minimal (no node_modules)
- Sufficient for a file upload form with navigation menu
- Modern look achieved with CSS (flexbox/grid, variables, modern selectors)
- Can use a lightweight CSS framework like Pico CSS (~10KB) for instant modern styling

**Alternatives considered**:
- React/Vue/Svelte SPA: Overkill for a file upload form, adds build complexity and image size
- HTMX: Could be useful but adds complexity without benefit for this simple use case
- Go HTML templates: Good for server-rendered pages, used for the confirmation page

## R4: MongoDB Driver

**Decision**: `go.mongodb.org/mongo-driver/v2` (official Go MongoDB driver)

**Rationale**:
- Official, well-maintained driver
- Supports all required operations: connect, insert, replace (upsert)
- Connection pooling built-in
- Context-based timeout support

## R5: HTTP Router

**Decision**: Go standard library `net/http` (Go 1.22+ with enhanced routing)

**Rationale**:
- Go 1.22+ supports method-based routing (`GET /path`, `POST /path`) natively
- No external dependency needed for this simple API (2-3 endpoints)
- Reduces binary size and dependency count

**Alternatives considered**:
- `chi`: Excellent router but unnecessary for 2-3 routes
- `gorilla/mux`: Archived project, not recommended for new projects
- `gin`/`echo`: Full frameworks, overkill for this use case

## R6: Docker Strategy

**Decision**: Multi-stage build — Go builder + `alpine` final image

**Rationale**:
- Stage 1: `golang:1.23-alpine` for compilation
- Stage 2: `alpine:3.19` (~7MB base) for runtime
- Go binary is statically compiled — only needs CA certificates for TLS
- Estimated final image: ~25-35MB

**Build**:
```dockerfile
FROM golang:1.23-alpine AS builder
# ... compile ...

FROM alpine:3.19
COPY --from=builder /app/smartlattes /smartlattes
EXPOSE 8080
CMD ["/smartlattes"]
```

## R7: JSON Conversion Strategy

**Decision**: Parse XML into Go structs, then serialize to JSON for MongoDB storage

**Rationale**:
- Lattes XML has a well-defined schema with known elements (CURRICULO-VITAE, DADOS-GERAIS, etc.)
- Use a generic XML-to-map approach to preserve all attributes and nested elements
- Store the full JSON document in MongoDB with `NUMERO-IDENTIFICADOR` as `_id`
- This preserves the complete XML structure without needing to model every field

**Approach**: Use a recursive XML-to-map converter that:
1. Reads XML elements and their attributes
2. Converts to nested `map[string]interface{}` structure
3. Attributes become fields with `@` prefix (e.g., `@NOME-COMPLETO`)
4. Child elements become nested maps or arrays
5. Store entire structure as a single MongoDB document
