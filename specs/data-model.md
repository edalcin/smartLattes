# Data Model: Lattes XML Upload

**Date**: 2026-02-07

## MongoDB Collections

### Collection: `curriculos`

Stores the complete Lattes CV data converted from XML to JSON.

**Identity**: `_id` = `NUMERO-IDENTIFICADOR` (string, e.g., `"8334174268306003"`)

**Document structure**:

```json
{
  "_id": "8334174268306003",
  "_metadata": {
    "uploadedAt": "2026-02-07T15:30:00Z",
    "originalFilename": "8334174268306003.xml",
    "fileSize": 311296
  },
  "CURRICULO-VITAE": {
    "@SISTEMA-ORIGEM-XML": "LATTES_OFFLINE",
    "@NUMERO-IDENTIFICADOR": "8334174268306003",
    "@DATA-ATUALIZACAO": "03012026",
    "@HORA-ATUALIZACAO": "200803",
    "DADOS-GERAIS": {
      "@NOME-COMPLETO": "Eduardo Couto Dalcin",
      "@NOME-EM-CITACOES-BIBLIOGRAFICAS": "DALCIN, E. C.;DALCIN, EDUARDO;...",
      "@NACIONALIDADE": "B",
      "RESUMO-CV": { ... },
      "ENDERECO": { ... },
      "FORMACAO-ACADEMICA-TITULACAO": { ... },
      "ATUACOES-PROFISSIONAIS": { ... },
      "AREAS-DE-ATUACAO": { ... },
      "IDIOMAS": { ... }
    },
    "PRODUCAO-BIBLIOGRAFICA": {
      "TRABALHOS-EM-EVENTOS": { ... },
      "ARTIGOS-PUBLICADOS": { ... },
      "LIVROS-E-CAPITULOS": { ... },
      "TEXTOS-EM-JORNAIS-OU-REVISTAS": { ... }
    },
    "PRODUCAO-TECNICA": {
      "SOFTWARE": [ ... ],
      "TRABALHO-TECNICO": [ ... ]
    },
    "OUTRA-PRODUCAO": { ... },
    "DADOS-COMPLEMENTARES": { ... }
  }
}
```

**Indexes**:
- `_id` (default unique index on NUMERO-IDENTIFICADOR)
- No additional indexes needed for the upload feature

**Conventions**:
- XML attributes are prefixed with `@` (e.g., `@NOME-COMPLETO`)
- XML child elements become nested objects or arrays
- When an element repeats (e.g., multiple `SOFTWARE` entries), they become an array
- `_metadata` field stores upload tracking info (not from the XML itself)

### Upsert Behavior

- **Insert**: If no document with the given `_id` exists, create it
- **Replace**: If a document with the given `_id` exists, replace it entirely with the new data (including updated `_metadata.uploadedAt`)
- MongoDB operation: `ReplaceOne` with `upsert: true`

### Data Lifecycle

- Documents are created on first upload
- Documents are fully replaced on re-upload (no partial updates)
- No deletion mechanism in the current feature scope
- No TTL or expiration policy
