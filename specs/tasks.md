# Tasks: smartLattes — Transformation Context

**Feature**: `smartlattes`
**Date**: 2026-02-08
**Plan**: [plan.md](plan.md) | **Spec**: [spec.md](spec.md)

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US3, US4, US5)
- Include exact file paths in descriptions

## Previously Completed (Acquisition Context)

User Stories 1 and 2 (Upload + Confirmation) are fully implemented:
- [x] T001-T025: Setup, foundational, upload, confirmation, navigation, and polish tasks

The following tasks cover the **Transformation context** (new) and navigation updates.

---

## Phase 1: Setup

**Purpose**: Add new dependencies, create the AI prompt file, and scaffold new packages.

- [ ] T101 [P] Create the AI summary prompt file at `resumoPrompt.md` in the repository root. The prompt must instruct the AI to generate in Portuguese: (a) analysis of researcher main strengths and characteristics, (b) structured areas of expertise, (c) scientific contribution potential, (d) most frequent co-authors with counts, (e) production quantified by area of expertise. The prompt must instruct the AI to output in Markdown format.
- [ ] T102 [P] Add Go dependencies for document export: `github.com/jung-kurt/gofpdf` (PDF) and `github.com/nguyenthenguyen/docx` (DOCX). Verify compatibility via Docker build (`golang:1.23-alpine`).
- [ ] T103 [P] Create `internal/ai/provider.go`: define `AIProvider` interface with `ListModels(ctx context.Context, apiKey string) ([]Model, error)` and `Generate(ctx context.Context, req GenerateRequest) (string, error)`. Define `Model` struct (`ID`, `DisplayName` string fields with json tags). Define `GenerateRequest` struct (`APIKey`, `Model`, `SystemPrompt`, `UserData` string; `MaxTokens` int). Implement `NewProvider(name string) (AIProvider, error)` factory that returns the correct provider for "openai", "anthropic", "gemini" or error for unknown.
- [ ] T104 [P] Create `internal/export/markdown.go`: implement `ToMarkdown(summary string) []byte` that returns the raw summary text as bytes with UTF-8 encoding.

---

## Phase 2: Foundational

**Purpose**: Core backend infrastructure that all transformation user stories depend on.

**Depends on**: Phase 1

### AI Provider Implementations (parallelizable)

- [ ] T105 [P] Implement OpenAI provider in `internal/ai/openai.go`. `ListModels`: `GET https://api.openai.com/v1/models` with `Authorization: Bearer KEY` header, filter to models owned by "openai" or "system" with id containing "gpt", return `[]Model`. `Generate`: `POST https://api.openai.com/v1/chat/completions` with JSON body `{model, messages: [{role:"system", content:prompt}, {role:"user", content:userData}]}`. Parse `choices[0].message.content`. Use 120s timeout context. Map HTTP 401/403 to auth errors.
- [ ] T106 [P] Implement Anthropic provider in `internal/ai/anthropic.go`. `ListModels`: `GET https://api.anthropic.com/v1/models` with headers `x-api-key: KEY` and `anthropic-version: 2023-06-01`, return `[]Model` from `data[]` array (use `id` and `display_name`). `Generate`: `POST https://api.anthropic.com/v1/messages` with JSON body `{model, max_tokens:4096, system:prompt, messages:[{role:"user", content:userData}]}`. Parse `content[0].text`. Use 120s timeout. Map HTTP 401/403 to auth errors.
- [ ] T107 [P] Implement Gemini provider in `internal/ai/gemini.go`. `ListModels`: `GET https://generativelanguage.googleapis.com/v1beta/models` with `x-goog-api-key: KEY` header, filter to models supporting "generateContent" in `supportedGenerationMethods`, return `[]Model` (use `name` stripped of "models/" prefix as ID, `displayName` as display name). `Generate`: `POST /v1beta/models/{model}:generateContent` with JSON body `{system_instruction:{parts:[{text:prompt}]}, contents:[{parts:[{text:userData}]}], generationConfig:{maxOutputTokens:4096}}`. Parse `candidates[0].content.parts[0].text`. Use 120s timeout.

### Token Truncation

- [ ] T108 Implement token truncation in `internal/ai/truncate.go`: function `TruncateCV(cvData map[string]interface{}, maxTokens int) (map[string]interface{}, bool)`. Deep-copy the input map. Serialize to JSON, estimate tokens as `len(json)/4`. If under limit, return original + false. Otherwise truncate progressively: (1) delete `dados-complementares`, (2) delete `outra-producao`, (3) delete `producao-tecnica`, (4) truncate arrays inside `producao-bibliografica` to first N items. Re-check after each step. Return truncated map + true.

### MongoDB Store Extensions (parallelizable)

- [ ] T109 [P] Add `SearchCVs(ctx context.Context, query string) ([]CVSummary, error)` to `internal/store/mongo.go`. Define `CVSummary` struct with `LattesID` and `Name` string fields. If query is all digits, search by `_id` prefix regex. Otherwise search by `curriculo-vitae.dados-gerais.nome-completo` with case-insensitive regex. Limit to 20 results. Project only `_id` and `curriculo-vitae.dados-gerais.nome-completo`.
- [ ] T110 [P] Add `GetCV(ctx context.Context, lattesID string) (map[string]interface{}, error)` to `internal/store/mongo.go`. Retrieve full document from `curriculos` collection by `_id`. Return nil + error if not found.
- [ ] T111 [P] Add `UpsertSummary(ctx context.Context, lattesID, summary, provider, model string) error` to `internal/store/mongo.go`. Upsert to `resumos` collection with `_id=lattesID`, `resumo=summary`, `_metadata.generatedAt=time.Now().UTC()`, `_metadata.provider=provider`, `_metadata.model=model`. Use `ReplaceOne` with `upsert:true`.
- [ ] T112 [P] Add `GetSummary(ctx context.Context, lattesID string) (*SummaryDoc, error)` to `internal/store/mongo.go`. Define `SummaryDoc` struct with `ID` (from `_id`), `Resumo` string, `Metadata` struct (`GeneratedAt` time.Time, `Provider` string, `Model` string). Retrieve from `resumos` collection by `_id`. Return nil + specific error if not found.

---

## Phase 3: US3 — Update Navigation Menu

**Goal**: Add "Gerar Resumo" to all page navbars and create the dedicated summary page with route.

**Depends on**: None (can run in parallel with Phase 2)
**Independent Test**: Navigate to homepage, verify 3 menu items ("Enviar CV", "Gerar Resumo", "Explorar Dados"). Click "Gerar Resumo" → page loads with search input.

- [ ] T113 [US3] Update `internal/static/index.html`: add "Gerar Resumo" menu item in the navbar between "Enviar CV" and "Explorar Dados", linking to `/resumo`. Add a third feature card on the homepage describing AI-powered researcher summary generation.
- [ ] T114 [P] [US3] Update `internal/static/upload.html`: add "Gerar Resumo" link to the navbar between existing items.
- [ ] T115 [P] [US3] Update `internal/static/explorer.html`: add "Gerar Resumo" link to the navbar between existing items.
- [ ] T116 [US3] Create `internal/static/resumo.html`: page with same navbar as other pages. Content: (a) search section with text input for lattesID/name and search results area, (b) AI config section (hidden initially) with provider pulldown (Gemini, OpenAI, Anthropic), API key password input, "Carregar Modelos" button, model pulldown (disabled), "Gerar Resumo" button (disabled), (c) summary display area (hidden), (d) download buttons (.md, .docx, .pdf) and "Ok" button (hidden). Use same CSS classes as upload.html.
- [ ] T117 [US3] Add route for `/resumo` in `internal/handler/pages.go`: add `PageHandler("resumo.html")` pattern.
- [ ] T118 [US3] Register `GET /resumo` route in `cmd/smartlattes/main.go` alongside existing page routes.

---

## Phase 4: US4 — Generate Researcher Summary via AI (P1)

**Goal**: Full end-to-end AI summary generation: model listing, generation, display, download (3 formats), and MongoDB storage.

**Depends on**: Phase 2 (AI providers, store extensions), Phase 3 (resumo page exists)
**Independent Test**: Upload CV XML → select provider → enter valid API key → load models → select model → click "Gerar Resumo" → verify summary has 5 elements from FR-104. Download .md, .docx, .pdf. Click "Ok" → verify document in `resumos` MongoDB collection.

### Backend Handlers

- [ ] T119 [US4] Create `internal/handler/search.go`: `SearchHandler` with `Store *store.MongoDB` field. Handle `GET /api/search?q=`. Validate `q` param exists and len >= 3 (else 400). Call `store.SearchCVs(ctx, q)`. Return JSON `{"success":true,"results":[{"lattesId":"...","name":"..."}]}`. Return 503 if MongoDB error.
- [ ] T120 [US4] Create `internal/handler/models.go`: `ModelsHandler` struct (no fields needed). Handle `POST /api/models`. Parse JSON body `{provider, apiKey}` (400 if missing). Call `ai.NewProvider(provider)` (400 if unknown provider). Call `provider.ListModels(ctx, apiKey)`. Return JSON `{"success":true,"models":[{"id":"...","displayName":"..."}]}`. Map auth errors to 401.
- [ ] T121 [US4] Create `internal/handler/summary.go`: `SummaryHandler` with `Store *store.MongoDB` and `Prompt string` fields. Add `//go:embed` for `resumoPrompt.md` (or receive prompt via constructor). Handle `POST /api/summary`: parse JSON `{lattesId, provider, apiKey, model}` (400 if missing). Get CV via `store.GetCV()` (404 if not found). Serialize CV to JSON string. Call `ai.TruncateCV()` if needed. Call `ai.NewProvider(provider).Generate()` with system=prompt, user=CV JSON. Return JSON `{"success":true,"summary":"...","truncated":false,"provider":"...","model":"..."}`. Include `truncationWarning` string if truncated. Map errors: auth→401, timeout→504, other→500.
- [ ] T122 [US4] Add save endpoint in `internal/handler/summary.go`: handle `POST /api/summary/save`. Parse JSON `{lattesId, summary, provider, model}` (400 if missing). Call `store.UpsertSummary()`. Return JSON `{"success":true,"message":"Resumo salvo com sucesso"}`.
- [ ] T123 [US4] Create `internal/handler/download.go`: `DownloadHandler` with `Store *store.MongoDB`. Handle `GET /api/download/{lattesID}`. Read `format` query param (required, must be md/docx/pdf, else 400). Get summary via `store.GetSummary()` (404 if not found). For `md`: set `Content-Type: text/markdown`, `Content-Disposition: attachment; filename=resumo-{lattesID}.md`, write raw text. For `pdf`: call `export.ToPDF()`, set PDF headers. For `docx`: call `export.ToDOCX()`, set DOCX headers.

### Export (parallelizable)

- [ ] T124 [P] [US4] Implement PDF export in `internal/export/pdf.go`: `ToPDF(summary, researcherName string) ([]byte, error)`. Use gofpdf: create A4 document, add UTF-8 font support, write title "Resumo do Pesquisador: {name}", write summary text as multi-line paragraphs. Return PDF bytes.
- [ ] T125 [P] [US4] Implement DOCX export in `internal/export/docx.go`: `ToDOCX(summary, researcherName string) ([]byte, error)`. Create a simple DOCX document with title and summary text. Return DOCX bytes.

### Route Registration

- [ ] T126 [US4] Register all new API routes in `cmd/smartlattes/main.go`: `GET /api/search` → `SearchHandler`, `POST /api/models` → `ModelsHandler`, `POST /api/summary` → `SummaryHandler`, `POST /api/summary/save` → `SummaryHandler`, `GET /api/download/` → `DownloadHandler`. Pass `Store` and prompt to handlers that need them.

### Frontend — Upload Page (inline summary after upload)

- [ ] T127 [US4] Update `internal/static/js/upload.js`: after successful upload confirmation, show an AI summary section. Add: provider pulldown (Gemini, OpenAI, Anthropic), API key password input, "Carregar Modelos" button. When clicked, call `POST /api/models` with `{provider, apiKey}`. On success, populate model pulldown and enable it. Show loading spinner during fetch. On error, show message inline.
- [ ] T128 [US4] Update `internal/static/js/upload.js`: add "Gerar Resumo" button (enabled only when model is selected). On click, call `POST /api/summary` with `{lattesId, provider, apiKey, model}` where lattesId comes from the upload response. Show loading spinner (can take up to 120s). On success, display summary Markdown in a styled container. Show truncation warning banner if `truncated:true`.
- [ ] T129 [US4] Update `internal/static/js/upload.js`: add download buttons (.md, .docx, .pdf) and "Ok" button below the summary. Downloads open `GET /api/download/{lattesId}?format=X` in new window. "Ok" calls `POST /api/summary/save` with `{lattesId, summary, provider, model}`, shows success toast, then allows starting over.

### Frontend — Dedicated Summary Page

- [ ] T130 [US4] Create `internal/static/js/resumo.js`: implement CV search. Add input event listener with 300ms debounce. When query >= 3 chars, call `GET /api/search?q=`. Display results as clickable cards showing name and lattesID. On card click, set selected CV (store lattesId and name) and reveal the AI config form. Show "Nenhum resultado encontrado" if empty results.
- [ ] T131 [US4] Extend `internal/static/js/resumo.js`: implement AI config form. Same logic as T127 — provider pulldown, API key, "Carregar Modelos" button, model pulldown. All reusing the same API calls.
- [ ] T132 [US4] Extend `internal/static/js/resumo.js`: implement summary generation, display, download, and save. Same logic as T128/T129 but using the lattesId from search selection instead of upload response.

### Styles

- [ ] T133 [US4] Update `internal/static/css/style.css`: add styles for AI summary components — provider/key/model form group, long-running loading spinner with message, summary result container (render Markdown headings, lists, bold), download buttons row (3 buttons side by side), truncation warning banner (yellow/amber), search results cards (clickable, hover state), search input with icon.

---

## Phase 5: US5 — Error Handling for Summary Generation (P2)

**Goal**: Robust error classification and user-friendly error display for all AI integration failure modes.

**Depends on**: Phase 4
**Independent Test**: Enter invalid API key → see "Chave de API invalida" message. Leave fields empty → validation messages. Simulate timeout → timeout message.

- [ ] T134 [US5] Define error types in `internal/ai/provider.go`: add sentinel errors `ErrInvalidKey`, `ErrProviderUnavailable`, `ErrTimeout`, `ErrRateLimited` using `errors.New()`. Update each provider (openai.go, anthropic.go, gemini.go) to return these errors by mapping HTTP status codes: 401/403 → `ErrInvalidKey`, 429 → `ErrRateLimited`, 5xx → `ErrProviderUnavailable`, `context.DeadlineExceeded` → `ErrTimeout`.
- [ ] T135 [US5] Update error handling in `internal/handler/models.go` and `internal/handler/summary.go`: use `errors.Is()` to map AI errors to HTTP responses. `ErrInvalidKey` → 401 "Chave de API invalida ou sem permissao para este provedor", `ErrProviderUnavailable` → 503 "Provedor de IA indisponivel. Tente novamente mais tarde.", `ErrTimeout` → 504 "Tempo limite excedido (120s). Tente um modelo menor ou tente novamente.", `ErrRateLimited` → 429 "Limite de requisicoes atingido. Aguarde e tente novamente."
- [ ] T136 [US5] Update frontend JS (both `upload.js` and `resumo.js`): display API error messages in a styled red banner. Add client-side validation: require provider selection, require API key (min 10 chars), require model selection before enabling "Gerar Resumo". Show specific user-friendly messages per HTTP status code (401, 429, 503, 504). Allow retry without re-entering API key and model selection.

---

## Phase 6: Polish & Cross-Cutting

**Purpose**: Final embed, Docker, end-to-end validation.

**Depends on**: Phase 5

- [ ] T137 Verify `resumoPrompt.md` embedding: ensure `//go:embed` directive in `internal/handler/summary.go` correctly loads the prompt at compile time. The embed path must be relative from the Go file's directory OR the prompt must be passed from `main.go` which embeds it from repo root. Confirm prompt is available without filesystem at runtime.
- [ ] T138 Update `Dockerfile`: ensure `resumoPrompt.md` is in the build context (COPY step). Verify build succeeds with new dependencies (gofpdf, docx). Verify final image size remains under 200MB. Test that container starts and all routes respond.
- [ ] T139 End-to-end validation: build Docker image, run with MongoDB, (1) upload `docs/8334174268306003.xml`, (2) verify no personal data in `curriculos` collection, (3) navigate to "Gerar Resumo" page, search for uploaded CV, (4) generate summary with at least one provider, (5) verify summary contains 5 elements from FR-104, (6) download in .md, .docx, .pdf, (7) click "Ok" and verify `resumos` collection has the document with correct metadata.

---

## Dependencies

```text
Phase 1 (Setup) ──┬── Phase 2 (Foundational) ──── Phase 4 (US4: AI Summary) ── Phase 5 (US5: Errors)
                   │                                       │                            │
                   └── Phase 3 (US3: Menu) ────────────────┘                    Phase 6 (Polish)
```

**Phase 3 can run in parallel with Phase 2** (no backend dependencies).

### Parallel Opportunities

| Phase | Parallel Tasks |
|-------|---------------|
| Phase 1 | T101, T102, T103, T104 — all independent |
| Phase 2 | T105, T106, T107 — three providers in parallel. T109-T112 — four store methods in parallel. |
| Phase 3 | T114, T115 — navbar updates in parallel. Entire phase parallel with Phase 2. |
| Phase 4 | T124, T125 — export formats in parallel. T119-T123 — handlers mostly parallel (different files). |

---

## Implementation Strategy

### MVP (minimum viable)

1. Phase 1 (T101-T104) → scaffold
2. Phase 2 with **one provider only** (T105 OpenAI + T108-T112) → core working
3. Phase 3 (T113-T118) → navigation
4. Phase 4 with .md export only (T119-T123, T126-T133 but skip T124/T125) → summary works end-to-end
5. **Validate**: upload → generate summary → display → download .md → save to MongoDB

### Increment 2

6. Add remaining providers (T106 Anthropic, T107 Gemini)
7. Add PDF and DOCX export (T124, T125)

### Increment 3

8. Phase 5 (T134-T136) → error handling polish
9. Phase 6 (T137-T139) → Docker and E2E verification

---

## Summary

| Phase | Tasks | Count | Description |
|-------|-------|-------|-------------|
| Phase 1 | T101-T104 | 4 | Setup: prompt, deps, package scaffolding |
| Phase 2 | T105-T112 | 8 | Foundational: 3 AI providers, truncation, store extensions |
| Phase 3 | T113-T118 | 6 | US3: Navigation update + resumo page |
| Phase 4 | T119-T133 | 15 | US4: AI summary (handlers, export, routes, frontend, styles) |
| Phase 5 | T134-T136 | 3 | US5: Error handling for AI integration |
| Phase 6 | T137-T139 | 3 | Polish: embed, Docker, E2E verification |
| **Total** | | **39** | |

---

## Notes

- Task IDs start at T101 to avoid collision with completed T001-T025
- All new code uses lowercase attribute keys (matching existing parser changes)
- API keys are NEVER persisted — used only for the duration of the HTTP request
- `resumoPrompt.md` is embedded at compile time via `//go:embed`
- Use `docs/8334174268306003.xml` as primary test artifact
- All UI text in Portuguese (Brazilian)
