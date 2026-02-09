# Quickstart: Contexto de Análise

**Date**: 2026-02-09

## Prerequisites

- Docker installed (all builds happen via Docker — no local Go required)
- MongoDB running at the URI specified in `.env`
- At least 2 CVs uploaded to the database (for analysis to work)
- Valid API key for at least one AI provider (OpenAI, Anthropic, or Gemini)

## Files to Create/Modify

### New Files

| File | Purpose |
|------|---------|
| `cmd/smartlattes/analisePrompt.md` | AI prompt for relationship analysis (embedded at compile time) |
| `internal/handler/analysis.go` | HTTP handler for analysis endpoints |
| `internal/static/analise.html` | Dedicated analysis page |
| `internal/static/js/analise.js` | Analysis page JavaScript logic |

### Modified Files

| File | Change |
|------|--------|
| `cmd/smartlattes/main.go` | Add `//go:embed analisePrompt.md`, register analysis routes, serve analise page |
| `internal/store/mongo.go` | Add `relacoes` collection methods: `UpsertAnalysis`, `GetAnalysis`, `GetAllCVSummaries`, `CountCVs` |
| `internal/ai/truncate.go` | Add `TruncateAnalysisData()` for multi-CV truncation |
| `internal/handler/download.go` | Add analysis download route (`/api/analysis/download/`) |
| `internal/static/index.html` | Add "Analisar Relações" menu item |
| `internal/static/upload.html` | Add analysis prompt section after summary |
| `internal/static/resumo.html` | Add analysis prompt section after summary |
| `internal/static/js/upload.js` | Add analysis flow after summary display |
| `internal/static/js/resumo.js` | Add analysis flow after summary display |
| `internal/static/css/style.css` | Add analysis-specific styles |
| `internal/static/embed.go` | No change needed (already embeds `*.html`, `js/`, `css/`) |

## Build and Test

```bash
# Build Docker image
docker build -t ghcr.io/edalcin/smartlattes . --progress=plain

# Run locally
docker run --rm -p 8080:8080 --env-file .env ghcr.io/edalcin/smartlattes

# Test flow:
# 1. Upload at least 2 CVs via /upload
# 2. Generate summary for one researcher (requires API key)
# 3. After summary, click "Sim" to analyze relationships
# 4. Verify two lists are generated
# 5. Test download in .md format
# 6. Verify data saved in MongoDB relacoes collection
```

## Key Patterns to Follow

1. **Handler pattern**: See `internal/handler/summary.go` for the exact pattern to replicate in `analysis.go`
2. **Store methods**: See `UpsertSummary`/`GetSummary` in `internal/store/mongo.go` for the pattern to replicate
3. **Frontend JS**: See `resumo.js` for the provider/model selection and generation flow to replicate
4. **Embed**: `analisePrompt.md` MUST be in `cmd/smartlattes/` (same directory as `main.go`) due to Go embed constraints
5. **Error sentinels**: Reuse `ai.ErrInvalidKey`, `ai.ErrTimeout`, etc. from `internal/ai/provider.go`
