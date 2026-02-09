# Research: smartLattes — Contexto de Análise

**Date**: 2026-02-09

## Research Tasks

### R1: How to retrieve and aggregate multiple CVs for AI analysis

**Decision**: Use MongoDB `Find()` with projection to retrieve all CVs except the current researcher's, projecting only analysis-relevant fields.

**Rationale**: The existing `store/mongo.go` already uses direct MongoDB driver v2 methods. A `Find()` with filter `{_id: {$ne: currentLattesID}}` and projection limits data transfer. This is simpler than aggregation pipelines and sufficient for the expected scale (tens to low hundreds of researchers).

**Alternatives considered**:
- MongoDB Aggregation Pipeline: Overkill for a simple filter+projection. Would add complexity without benefit.
- In-memory filtering: Retrieve all, filter in Go. Wasteful for large collections — rejected.
- Pre-computed summaries: Store extracted keywords per CV. Too much infrastructure for the current scope.

### R2: Token management for multi-CV payloads

**Decision**: Extend existing `TruncateCV()` pattern with a new `TruncateAnalysisData()` function that handles multiple CVs. Strategy: always include full current CV, progressively reduce other CVs' data.

**Rationale**: The existing `ai/truncate.go` already handles single-CV truncation with a ~4 chars/token estimate. The same approach scales to multiple CVs with a priority-based reduction strategy.

**Alternatives considered**:
- Chunking (multiple API calls): Would require merging results from multiple AI calls. Inconsistent results, higher cost, more complex error handling — rejected.
- Embedding-based relevance: Use vector embeddings to select most relevant CVs first. Requires embedding infrastructure not in scope — rejected for now.
- Fixed limit on researchers: Hard cap at N researchers. Too restrictive and arbitrary — rejected in favor of dynamic truncation.

### R3: Prompt design for relationship analysis

**Decision**: Single prompt in `analisePrompt.md` that receives the current researcher's full data plus summarized data of all other researchers. The AI generates a structured Markdown response with two clearly separated sections.

**Rationale**: Follows the exact same pattern as `resumoPrompt.md` — a single system prompt with structured output requirements. The AI can handle multi-researcher comparison in a single call given sufficient context window.

**Alternatives considered**:
- Two-step analysis (first identify areas, then match): More API calls, higher latency, no clear quality improvement — rejected.
- Separate prompts for common vs complementary: Redundant processing of the same data — rejected.

### R4: Reusing AI credentials from summary step

**Decision**: Frontend passes the same `provider`, `apiKey`, and `model` values to the analysis endpoint. No server-side session storage of credentials.

**Rationale**: The spec requires reusing credentials without re-asking (FR-207). Since the frontend already holds these values in JavaScript variables (`currentProvider`, `currentApiKey`, `currentModel`), they can be passed directly to the analysis API call. This maintains the stateless server architecture and the "no credential storage" principle (FR-107).

**Alternatives considered**:
- Server-side session to hold credentials: Violates the "no credential storage" constraint — rejected.
- Re-ask the user: Violates FR-207 requirement — rejected.

### R5: Analysis storage schema

**Decision**: New `relacoes` collection with `_id` = lattesID, containing the full Markdown analysis text and metadata (provider, model, timestamp, count of researchers analyzed). Upsert behavior matches existing `resumos` pattern.

**Rationale**: Follows the exact same pattern as the `resumos` collection. The `researchersAnalyzed` count in metadata is useful for transparency (user knows how many researchers were compared).

**Alternatives considered**:
- Structured storage (separate arrays for common/complementary): Would require parsing AI output. Fragile and unnecessary — the Markdown is the primary format.
- Store in same `resumos` collection: Conflates two different artifact types — rejected for clarity.

### R6: Dedicated analysis page vs inline-only

**Decision**: Both. Analysis is available inline (after summary in upload/resumo pages) AND as a dedicated page accessible from the menu ("Analisar Relações").

**Rationale**: The spec defines a menu item "Analisar Relações" (FR-007, User Story 3 scenario 4) and the analysis flow after summary (FR-200). Supporting both entry points provides flexibility. The dedicated page reuses the same search + provider selection pattern from `resumo.html`.

**Alternatives considered**:
- Inline only (no dedicated page): Would leave the menu item as a placeholder. Inconsistent with spec — rejected.
- Dedicated page only: Would miss the natural flow after summary generation — rejected.
