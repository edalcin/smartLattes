# Data Model: smartLattes

**Date**: 2026-02-08

## MongoDB Collections

### Collection: `curriculos`

Stores the Lattes CV data converted from XML to JSON. All attribute keys are stored in lowercase.

**Identity**: `_id` = `numero-identificador` (string, e.g., `"8334174268306003"`)

**Document structure**:

```json
{
  "_id": "8334174268306003",
  "_metadata": {
    "uploadedAt": "2026-02-07T15:30:00Z",
    "originalFilename": "8334174268306003.xml",
    "fileSize": 311296
  },
  "curriculo-vitae": {
    "sistema-origem-xml": "LATTES_OFFLINE",
    "numero-identificador": "8334174268306003",
    "data-atualizacao": "03012026",
    "hora-atualizacao": "200803",
    "dados-gerais": {
      "nome-completo": "Eduardo Couto Dalcin",
      "orcid-id": "0000-0001-5000-0000",
      "nome-em-citacoes-bibliograficas": "DALCIN, E. C.;DALCIN, EDUARDO;...",
      "formacao-academica-titulacao": { ... },
      "atuacoes-profissionais": { ... },
      "areas-de-atuacao": { ... }
    },
    "producao-bibliografica": {
      "trabalhos-em-eventos": { ... },
      "artigos-publicados": { ... },
      "livros-e-capitulos": { ... },
      "textos-em-jornais-ou-revistas": { ... }
    },
    "producao-tecnica": {
      "software": [ ... ],
      "trabalho-tecnico": [ ... ]
    },
    "outra-producao": { ... },
    "dados-complementares": { ... }
  }
}
```

**Filtered fields in `dados-gerais`**: Only the following keys are stored (all others are discarded to exclude personal/sensitive data):
- `nome-completo`
- `orcid-id`
- `nome-em-citacoes-bibliograficas`
- `formacao-academica-titulacao`
- `atuacoes-profissionais`
- `areas-de-atuacao`

**Discarded data** (never stored): CPF, RG, passport, nationality, gender, race, address, phone, etc.

**Indexes**:
- `_id` (default unique index on numero-identificador)

**Conventions**:
- All keys are stored in lowercase (e.g., `nome-completo`, `producao-bibliografica`)
- XML child elements become nested objects or arrays
- When an element repeats (e.g., multiple `software` entries), they become an array
- `_metadata` field stores upload tracking info (not from the XML itself)

### Collection: `resumos`

Stores AI-generated researcher summaries (transformation context).

**Identity**: `_id` = `lattesID` (string, same as `numero-identificador` from `curriculos`)

**Document structure**:

```json
{
  "_id": "8334174268306003",
  "_metadata": {
    "generatedAt": "2026-02-08T10:30:00Z",
    "provider": "anthropic",
    "model": "claude-sonnet-4-5-20250929"
  },
  "resumo": "# Resumo do Pesquisador\n\n..."
}
```

**Fields**:
- `_id`: lattesID do pesquisador (chave primaria, mesma de `curriculos`)
- `_metadata.generatedAt`: timestamp UTC da geracao
- `_metadata.provider`: provedor de IA utilizado (`gemini`, `openai`, `anthropic`)
- `_metadata.model`: modelo especifico utilizado
- `resumo`: texto do resumo gerado em formato Markdown

**Indexes**:
- `_id` (default unique index on lattesID)

### Upsert Behavior

Applies to both `curriculos` and `resumos` collections:

- **Insert**: If no document with the given `_id` exists, create it
- **Replace**: If a document with the given `_id` exists, replace it entirely
- MongoDB operation: `ReplaceOne` with `upsert: true`

### Data Lifecycle

- `curriculos`: Documents are created on first upload and fully replaced on re-upload
- `resumos`: Documents are created on first generation and replaced on re-generation
- No deletion mechanism in the current feature scope
- No TTL or expiration policy
