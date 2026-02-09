# Data Model: smartLattes — Contexto de Análise

**Date**: 2026-02-09

## New Collection: `relacoes`

Stores AI-generated researcher relationship analysis (analysis context).

**Identity**: `_id` = `lattesID` (string, same as `numero-identificador` from `curriculos`)

**Document structure**:

```json
{
  "_id": "8334174268306003",
  "_metadata": {
    "generatedAt": "2026-02-09T14:00:00Z",
    "provider": "anthropic",
    "model": "claude-sonnet-4-5-20250929",
    "researchersAnalyzed": 15
  },
  "analise": "# Análise de Relações do Pesquisador\n\n## Pesquisadores com Interesses Comuns\n\n### 1. Dr. Maria Silva (LattesID: 1234567890123456)\n**Áreas em comum**: Biodiversidade, Taxonomia, Conservação\n...\n\n## Pesquisadores com Interesses Complementares\n\n### 1. Dr. João Souza (LattesID: 9876543210987654)\n**Áreas complementares**: Bioinformática × Taxonomia\n**Sugestão de projeto**: ...\n..."
}
```

**Fields**:
- `_id`: lattesID do pesquisador (chave primária, mesma de `curriculos`)
- `_metadata.generatedAt`: timestamp UTC da geração
- `_metadata.provider`: provedor de IA utilizado (`gemini`, `openai`, `anthropic`)
- `_metadata.model`: modelo específico utilizado
- `_metadata.researchersAnalyzed`: quantidade de pesquisadores comparados na análise
- `analise`: texto da análise gerada em formato Markdown, contendo duas seções:
  - "Pesquisadores com Interesses Comuns"
  - "Pesquisadores com Interesses Complementares"

**Indexes**:
- `_id` (default unique index on lattesID)

### Upsert Behavior

Same pattern as `curriculos` and `resumos`:
- **Insert**: If no document with the given `_id` exists, create it
- **Replace**: If a document with the given `_id` exists, replace it entirely (re-analysis overwrites previous)
- MongoDB operation: `ReplaceOne` with `upsert: true`

### Data Lifecycle

- Documents are created on first analysis and replaced on re-analysis
- No deletion mechanism in the current feature scope
- No TTL or expiration policy

## New Store Methods

```go
// UpsertAnalysis stores or replaces the analysis for a researcher
UpsertAnalysis(ctx context.Context, lattesID string, analysis string, provider string, model string, researchersAnalyzed int) error

// GetAnalysis retrieves the stored analysis for a researcher
GetAnalysis(ctx context.Context, lattesID string) (*AnalysisDoc, error)

// GetAllCVSummaries retrieves all CVs except the specified one, with projection
GetAllCVSummaries(ctx context.Context, excludeLattesID string) ([]map[string]interface{}, error)

// CountCVs returns the total number of CVs in the database
CountCVs(ctx context.Context) (int64, error)
```

### AnalysisDoc Type

```go
type AnalysisDoc struct {
    ID       string          `bson:"_id"`
    Analise  string          `bson:"analise"`
    Metadata AnalysisMetadata `bson:"_metadata"`
}

type AnalysisMetadata struct {
    GeneratedAt         time.Time `bson:"generatedAt"`
    Provider            string    `bson:"provider"`
    Model               string    `bson:"model"`
    ResearchersAnalyzed int       `bson:"researchersAnalyzed"`
}
```

## Projection for CV Retrieval (Analysis)

When retrieving CVs for analysis, use projection to minimize data:

```go
projection := bson.M{
    "_id": 1,
    "curriculo-vitae.dados-gerais.nome-completo": 1,
    "curriculo-vitae.dados-gerais.areas-de-atuacao": 1,
    "curriculo-vitae.dados-gerais.formacao-academica-titulacao": 1,
    "curriculo-vitae.dados-gerais.atuacoes-profissionais": 1,
    "curriculo-vitae.producao-bibliografica": 1,
}
```

This projection excludes `producao-tecnica`, `outra-producao`, `dados-complementares`, and `_metadata` to reduce token usage.
