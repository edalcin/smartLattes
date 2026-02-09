# Tasks: smartLattes — Analysis Context

**Feature**: `smartlattes` — Contexto de Análise
**Date**: 2026-02-09
**Plan**: [plan.md](plan.md) | **Spec**: [../spec.md](../spec.md)

## Format: `[ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US6, US7)
- Include exact file paths in descriptions

## Previously Completed

- [x] T001-T025: Acquisition context (US1, US2 — upload + confirmation)
- [x] T101-T139: Transformation context (US3, US4, US5 — navigation, AI summary, error handling)

The following tasks cover the **Analysis context** (new): User Stories 6 and 7 (FR-200 to FR-209).

---

## Phase 1: Setup

**Purpose**: Create the analysis prompt file and verify existing infrastructure supports the new feature.

- [x] T201 [P] Create the AI analysis prompt file at `cmd/smartlattes/analisePrompt.md`. The prompt must instruct the AI to analyze a target researcher's CV data against all other researchers provided, and generate two Markdown sections in Portuguese (BR): (a) "Pesquisadores com Interesses Comuns" — for each match list name, lattesID, common areas of expertise, shared production themes, and why they would form a good research group; (b) "Pesquisadores com Interesses Complementares" — for each match list name, lattesID, complementary areas, and a concrete interdisciplinary project suggestion combining their distinct skills to address complex problems. The prompt must require data-driven analysis citing specific areas and publications, no speculation.
- [x] T202 [P] Verify existing `internal/ai/provider.go` interface and sentinel errors (`ErrInvalidKey`, `ErrTimeout`, `ErrRateLimited`, `ErrProviderUnavailable`) are sufficient for the analysis feature. No changes expected — this is a verification task.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Backend data layer and truncation logic that MUST be complete before analysis endpoints can work.

**Depends on**: Phase 1

### MongoDB Store Extensions (parallelizable)

- [x] T203 [P] Add `AnalysisDoc` and `AnalysisMetadata` types to `internal/store/mongo.go`: `AnalysisDoc` with fields `ID string` (bson `_id`), `Analise string` (bson `analise`), `Metadata AnalysisMetadata` (bson `_metadata`). `AnalysisMetadata` with fields `GeneratedAt time.Time`, `Provider string`, `Model string`, `ResearchersAnalyzed int`.
- [x] T204 [P] Add `CountCVs(ctx context.Context) (int64, error)` to `internal/store/mongo.go`. Use `curriculos` collection `CountDocuments()` with empty filter. Returns total number of CVs in database.
- [x] T205 [P] Add `GetAllCVSummaries(ctx context.Context, excludeLattesID string) ([]map[string]interface{}, error)` to `internal/store/mongo.go`. Use `Find()` on `curriculos` collection with filter `{_id: {$ne: excludeLattesID}}`. Use projection to include only: `_id`, `curriculo-vitae.dados-gerais.nome-completo`, `curriculo-vitae.dados-gerais.areas-de-atuacao`, `curriculo-vitae.dados-gerais.formacao-academica-titulacao`, `curriculo-vitae.dados-gerais.atuacoes-profissionais`, `curriculo-vitae.producao-bibliografica`. Decode all results into `[]map[string]interface{}` and return.
- [x] T206 [P] Add `UpsertAnalysis(ctx context.Context, lattesID, analysis, provider, model string, researchersAnalyzed int) error` to `internal/store/mongo.go`. Upsert to `relacoes` collection with `_id=lattesID`, `analise=analysis`, `_metadata.generatedAt=time.Now().UTC()`, `_metadata.provider=provider`, `_metadata.model=model`, `_metadata.researchersAnalyzed=researchersAnalyzed`. Use `ReplaceOne` with `upsert:true`.
- [x] T207 [P] Add `GetAnalysis(ctx context.Context, lattesID string) (*AnalysisDoc, error)` to `internal/store/mongo.go`. Retrieve from `relacoes` collection by `_id`. Return `nil` + `mongo.ErrNoDocuments` if not found.

### Token Truncation for Multi-CV Analysis

- [x] T208 Add `TruncateAnalysisData(currentCV map[string]interface{}, otherCVs []map[string]interface{}, maxTokens int) (string, bool)` to `internal/ai/truncate.go`. Build a combined JSON structure: `{"pesquisador_alvo": currentCV, "outros_pesquisadores": otherCVs}`. Estimate tokens as `len(json)/4`. If under limit, return JSON string + false. If over limit, progressively reduce other CVs: (1) remove `producao-bibliografica` from each other CV, (2) remove `atuacoes-profissionais` from each other CV, (3) limit number of other CVs (keep first N that fit). Always preserve full `currentCV`. Return truncated JSON string + true.

**Checkpoint**: Backend data layer ready — analysis endpoints can now be built.

---

## Phase 3: US3 — Update Navigation Menu (Priority: P3)

**Goal**: Add "Analisar Relações" menu item to all pages and create the dedicated analysis page with route.

**Depends on**: None (can run in parallel with Phase 2)
**Independent Test**: Navigate to homepage → verify 4 menu items ("Enviar CV", "Gerar Resumo", "Analisar Relações", "Explorar Dados"). Click "Analisar Relações" → page loads.

- [x] T209 [US3] Update `internal/static/index.html`: add "Analisar Relações" menu item in the navbar between "Gerar Resumo" and "Explorar Dados", linking to `/analise`. Add a feature card on the homepage describing AI-powered researcher relationship analysis.
- [x] T210 [P] [US3] Update `internal/static/upload.html`: add "Analisar Relações" link to the navbar between "Gerar Resumo" and "Explorar Dados".
- [x] T211 [P] [US3] Update `internal/static/resumo.html`: add "Analisar Relações" link to the navbar between "Gerar Resumo" and "Explorar Dados".
- [x] T212 [P] [US3] Update `internal/static/explorer.html`: add "Analisar Relações" link to the navbar between "Gerar Resumo" and "Explorar Dados".
- [x] T213 [US3] Create `internal/static/analise.html`: page with same navbar as other pages (including all 4 menu items). Content: (a) search section with text input for lattesID/name and search results area (same pattern as resumo.html), (b) AI config section (hidden initially) with provider pulldown (Gemini, OpenAI, Anthropic), API key password input, "Carregar Modelos" button, model pulldown (disabled), "Analisar Relações" button (disabled), (c) analysis results display area (hidden), (d) download buttons (.md, .docx, .pdf) and "Ok" button (hidden). Use same CSS classes as existing pages. All text in Portuguese (BR).
- [x] T214 [US3] Add route for `/analise` in `cmd/smartlattes/main.go`: register `PageHandler("analise.html")` alongside existing page routes.

**Checkpoint**: All pages show 4 menu items, dedicated analysis page loads.

---

## Phase 4: US6 — Analyze Researcher Relationships via AI (Priority: P2)

**Goal**: Full end-to-end relationship analysis: retrieve all CVs, send to AI with analysis prompt, display two lists (common + complementary interests), download in 3 formats, persist to MongoDB `relacoes` collection.

**Depends on**: Phase 2 (store extensions, truncation), Phase 3 (analise page exists)
**Independent Test**: With at least 2 CVs in database → generate summary → click "Sim, analisar" → verify two lists appear (common interests, complementary interests with project suggestions). Download .md. Click "Ok" → verify document in `relacoes` collection with correct metadata.

### Backend Handler

- [x] T215 [US6] Create `internal/handler/analysis.go`: `AnalysisHandler` struct with `Store *store.MongoDB` and `Prompt string` fields. Implement `ServeHTTP` that routes to `handleGenerate()` for `POST /api/analysis`, `handleSave()` for `POST /api/analysis/save`.
- [x] T216 [US6] Implement `handleGenerate()` in `internal/handler/analysis.go`: parse JSON `{lattesId, provider, apiKey, model}` (400 if missing). Get current CV via `store.GetCV()` (404 if not found). Call `store.CountCVs()` — if count <= 1, return 409 with `{"success":false,"error":"Não há outros pesquisadores na base para comparação"}` (FR-206). Call `store.GetAllCVSummaries(ctx, lattesId)` to get all other CVs. Call `ai.TruncateAnalysisData(currentCV, otherCVs, 100000)`. Call `ai.NewProvider(provider).Generate()` with system=prompt, user=truncatedData. Return JSON `{"success":true,"analysis":"...","provider":"...","model":"...","researchersAnalyzed":N,"truncated":bool,"truncationWarning":"..."}`. Map errors: `ErrInvalidKey`→401, `ErrTimeout`→504, `ErrProviderUnavailable`→503, MongoDB errors→503.
- [x] T217 [US6] Implement `handleSave()` in `internal/handler/analysis.go`: parse JSON `{lattesId, analysis, provider, model, researchersAnalyzed}` (400 if missing). Call `store.UpsertAnalysis()`. Return JSON `{"success":true,"message":"Análise salva com sucesso"}`.

### Download Handler Extension

- [x] T218 [US6] Create `internal/handler/analysis_download.go` (or extend `download.go`): handle `GET /api/analysis/download/{lattesId}?format=md|docx|pdf`. Get analysis via `store.GetAnalysis()` (404 if not found). For `md`: set `Content-Type: text/markdown`, `Content-Disposition: attachment; filename=analise-{lattesId}.md`, write raw text. For `docx` and `pdf`: return 501 Not Implemented (same pattern as existing summary download stubs).

### Route Registration

- [x] T219 [US6] Register analysis routes in `cmd/smartlattes/main.go`: add `//go:embed analisePrompt.md` directive and `var analisePrompt string`. Create `AnalysisHandler` with store and embedded prompt. Register `POST /api/analysis` → `AnalysisHandler`, `POST /api/analysis/save` → `AnalysisHandler`, `GET /api/analysis/download/` → analysis download handler. Pass store and prompt.

### Frontend — Inline Analysis After Summary (upload page)

- [x] T220 [US6] Update `internal/static/upload.html`: add a hidden "analysis prompt" section after the summary display area. Contains: card with text "Deseja analisar relações com outros pesquisadores?", two buttons "Sim, analisar" and "Não, obrigado", analysis results container (hidden), analysis download buttons (.md, .docx, .pdf) + "Ok" button (hidden), progress spinner area.
- [x] T221 [US6] Update `internal/static/js/upload.js`: after summary is displayed and saved, reveal the analysis prompt section. Store `currentProvider`, `currentApiKey`, `currentModel` from the summary step. On "Sim, analisar" click: show spinner with "Analisando relações entre pesquisadores..." (FR-209), call `POST /api/analysis` with `{lattesId: currentLattesId, provider: currentProvider, apiKey: currentApiKey, model: currentModel}`. On success: hide prompt, show analysis results (render Markdown with existing renderer), show download buttons + "Ok". On error: show error message in red banner, allow retry. On "Não, obrigado": hide the analysis prompt section entirely.
- [x] T222 [US6] Extend `internal/static/js/upload.js`: wire download buttons to `GET /api/analysis/download/{lattesId}?format=md|docx|pdf` (open in new window). Wire "Ok" button to call `POST /api/analysis/save` with `{lattesId, analysis: currentAnalysis, provider, model, researchersAnalyzed}`, show success toast.

### Frontend — Inline Analysis After Summary (resumo page)

- [x] T223 [US6] Update `internal/static/resumo.html`: add same hidden analysis prompt section as in upload.html (card with question, buttons, results area, download buttons, progress spinner).
- [x] T224 [US6] Update `internal/static/js/resumo.js`: same logic as T221 — after summary display, reveal analysis prompt, handle "Sim"/"Não" clicks, call `/api/analysis`, display results, handle errors.
- [x] T225 [US6] Extend `internal/static/js/resumo.js`: same logic as T222 — wire download and "Ok" buttons for the analysis results.

### Frontend — Dedicated Analysis Page

- [x] T226 [US6] Create `internal/static/js/analise.js`: implement CV search with debounce (same pattern as resumo.js). On card click, set selected CV and reveal AI config form. Implement provider pulldown, API key input, "Carregar Modelos" button calling `POST /api/models`, model pulldown. Add "Analisar Relações" button that calls `POST /api/analysis` with selected lattesId + provider + apiKey + model. Show spinner with "Analisando relações entre pesquisadores..." during call. On success: display analysis Markdown, show download buttons + "Ok". On error: show error message, allow retry. Wire download buttons to `GET /api/analysis/download/{lattesId}?format=`. Wire "Ok" to `POST /api/analysis/save`.

### Styles

- [x] T227 [US6] Update `internal/static/css/style.css`: add styles for analysis components — analysis prompt card (centered question with two action buttons), analysis results container (same styling as summary results), analysis progress spinner with custom message text, "Sim/Não" button pair styling (primary green for "Sim", neutral gray for "Não").

**Checkpoint**: Full analysis flow works — upload CV → generate summary → analyze relationships → see two lists → download .md → save to MongoDB.

---

## Phase 5: US7 — Error Handling for Relationship Analysis (Priority: P3)

**Goal**: Robust error handling specific to the analysis feature: single researcher in base, token overflow, provider errors, MongoDB unavailability.

**Depends on**: Phase 4
**Independent Test**: (1) With only 1 CV in database, trigger analysis → see "Não há outros pesquisadores" message. (2) Verify truncation warning appears when data exceeds limits. (3) Verify provider error messages display correctly.

- [x] T228 [US7] Update `internal/handler/analysis.go` `handleGenerate()`: ensure all error paths return clear Portuguese messages. Verify: `CountCVs <= 1` → 409 "Não há outros pesquisadores na base para comparação" (FR-206). `ErrInvalidKey` → 401 "Chave de API inválida ou sem permissão". `ErrTimeout` → 504 "Tempo limite excedido na análise de relações (120s)". `ErrProviderUnavailable` → 503 "Provedor de IA indisponível". MongoDB connection error during `GetAllCVSummaries` → 503 "Banco de dados indisponível". Add truncation warning text: "Os dados foram reduzidos para caber no limite do modelo. Alguns pesquisadores podem não ter sido incluídos na análise." (FR-208).
- [x] T229 [US7] Update frontend JS (all three: `upload.js`, `resumo.js`, `analise.js`): handle analysis-specific HTTP error codes. 409 → show info message (blue banner) "Não há outros pesquisadores na base para comparação". 401 → red banner with auth error. 503 → red banner with service unavailable. 504 → red banner with timeout message. For all errors: allow retry with same credentials. Show truncation warning (amber banner) when response has `truncated: true`.
- [x] T230 [US7] Update `internal/static/css/style.css`: add info banner style (blue, for the single-researcher case) to complement existing error (red) and warning (amber) banner styles.

**Checkpoint**: All error scenarios handled gracefully with clear Portuguese messages.

---

## Phase 6: Polish & Cross-Cutting

**Purpose**: Embed verification, Docker build, end-to-end validation.

**Depends on**: Phase 5

- [x] T231 Verify `analisePrompt.md` embedding: ensure `//go:embed analisePrompt.md` directive in `cmd/smartlattes/main.go` correctly loads the prompt at compile time. Confirm the embedded string is passed to `AnalysisHandler`. Verify prompt is available without filesystem at runtime.
- [x] T232 Update `Dockerfile`: verify build succeeds with new files (`analisePrompt.md`, `analysis.go`, `analise.html`, `analise.js`). Verify all embedded files are in the build context. Verify final image size remains under 200MB. Test that container starts and all new routes respond.
- [x] T233 End-to-end validation: build Docker image, run with MongoDB. (1) Upload at least 2 different CVs via `/upload`. (2) Generate summary for one researcher. (3) Click "Sim, analisar" → verify two lists appear. (4) Verify "Pesquisadores com Interesses Comuns" section exists. (5) Verify "Pesquisadores com Interesses Complementares" section exists with project suggestions. (6) Download in .md format. (7) Click "Ok" → verify `relacoes` collection has document with `_id` = lattesID, `analise` field, and `_metadata` with `researchersAnalyzed` count. (8) Navigate to "Analisar Relações" page → search → run analysis independently. (9) With only 1 CV, verify single-researcher message.

---

## Dependencies

```text
Phase 1 (Setup) ──┬── Phase 2 (Foundational) ──── Phase 4 (US6: Analysis) ── Phase 5 (US7: Errors)
                   │                                       │                            │
                   └── Phase 3 (US3: Menu) ────────────────┘                    Phase 6 (Polish)
```

**Phase 3 can run in parallel with Phase 2** (no backend dependencies).

### Parallel Opportunities

| Phase | Parallel Tasks |
|-------|---------------|
| Phase 1 | T201, T202 — independent |
| Phase 2 | T203-T207 — five store methods in parallel. T208 depends on no store tasks. |
| Phase 3 | T210, T211, T212 — navbar updates in parallel. Entire phase parallel with Phase 2. |
| Phase 4 | T215-T219 — backend tasks mostly sequential. T220-T227 — frontend tasks partially parallel (upload vs resumo vs analise pages). |

---

## Implementation Strategy

### MVP (minimum viable)

1. Phase 1 (T201-T202) → prompt + verification
2. Phase 2 with store methods (T203-T208) → data layer ready
3. Phase 3 (T209-T214) → navigation + analise page
4. Phase 4 backend only (T215-T219) → API endpoints working
5. Phase 4 inline on upload page (T220-T222) → analysis flow after summary
6. **Validate**: upload 2 CVs → generate summary → click "Sim" → see two lists → download .md → save to MongoDB

### Increment 2

7. Phase 4 inline on resumo page (T223-T225) → analysis also available from dedicated summary page
8. Phase 4 dedicated analise page (T226) → standalone analysis via menu
9. Phase 4 styles (T227) → visual polish

### Increment 3

10. Phase 5 (T228-T230) → error handling polish
11. Phase 6 (T231-T233) → Docker and E2E verification

---

## Summary

| Phase | Tasks | Count | Description |
|-------|-------|-------|-------------|
| Phase 1 | T201-T202 | 2 | Setup: analysis prompt, infrastructure verification |
| Phase 2 | T203-T208 | 6 | Foundational: store methods (relacoes), multi-CV truncation |
| Phase 3 | T209-T214 | 6 | US3: Navigation update + analise page |
| Phase 4 | T215-T227 | 13 | US6: Analysis (handler, download, routes, frontend ×3, styles) |
| Phase 5 | T228-T230 | 3 | US7: Error handling for analysis |
| Phase 6 | T231-T233 | 3 | Polish: embed, Docker, E2E verification |
| **Total** | | **33** | |

---

## Notes

- Task IDs start at T201 to follow the FR-200+ convention and avoid collision with T101-T139
- The analysis reuses the existing AI provider interface — no changes to `provider.go`, `openai.go`, `anthropic.go`, `gemini.go`
- `analisePrompt.md` MUST be in `cmd/smartlattes/` (not repo root) for `//go:embed` compatibility
- Provider/model/apiKey are reused from the summary step (FR-207) — no re-prompting
- The `relacoes` collection follows the exact same upsert pattern as `resumos`
- All UI text in Portuguese (Brazilian)
- DOCX/PDF download stubs return 501 (same as existing summary export stubs)
- Use at least 2 different CVs in the database for meaningful analysis testing
