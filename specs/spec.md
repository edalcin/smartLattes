# Feature Specification: smartLattes

**Feature**: `smartlattes`
**Created**: 2026-02-07
**Updated**: 2026-02-08
**Status**: Draft

## Visao Geral

O smartLattes ingere curriculos exportados em XML da Plataforma Lattes (CNPq), armazena-os em MongoDB e, a partir desses dados, gera artefatos de inteligencia por meio de provedores de IA (contexto de Transformacao). A arquitetura segue o C4 Model com tres contextos:

1. **Aquisicao** - Upload e armazenamento de curriculos Lattes (implementado).
2. **Transformacao** - Geracao de artefatos derivados usando IA (resumo do pesquisador, analises de relacao entre pesquisadores, perfil de pontos fortes).
3. **Apresentacao** - Visualizacao e exploracao dos dados na base (futuro).

## Clarifications

### Session 2026-02-07

- Q: Does the system require authentication or access control? → A: No authentication — open access, intended for use on a trusted/internal network.
- Q: What is the maximum upload file size? → A: 10MB maximum per uploaded XML file.
- Q: What should happen when MongoDB is unavailable at upload time? → A: Fail immediately with a clear error message; the user retries manually.

### Session 2026-02-08

- Dados pessoais e sensiveis (CPF, RG, raca, cor, endereco, telefone, etc.) NAO sao armazenados. De DADOS-GERAIS, apenas: nome-completo, orcid-id, nome-em-citacoes-bibliograficas, formacao-academica-titulacao, atuacoes-profissionais, areas-de-atuacao.
- Todos os atributos sao armazenados em caixa baixa no MongoDB (ex.: `areas-de-atuacao`).
- O contexto de Transformacao e introduzido com o primeiro componente: geracao de resumo do pesquisador via IA.
- A chave de API do provedor de IA e informada pelo usuario a cada uso; NAO e armazenada no servidor.
- O prompt de geracao do resumo fica em arquivo editavel no repositorio (`resumoPrompt.md`).
- Q: Como selecionar o modelo de IA? → A: O usuario escolhe o modelo em um segundo pulldown, apos selecionar o provedor. O sistema lista os modelos disponiveis via API do provedor usando a chave fornecida.
- Q: Quando a geracao de resumo esta disponivel? → A: Tanto apos o upload quanto para qualquer CV ja armazenado, via pagina dedicada acessivel pelo menu.
- Q: HTTPS ou HTTP para proteger chave de API em transito? → A: Plain HTTP — rede interna confiavel, sem necessidade de TLS.
- Q: Como buscar CV na pagina dedicada de resumo? → A: Campo de texto que busca por lattesID ou nome do pesquisador.
- Q: O que fazer quando os dados do CV excedem o limite de tokens do modelo? → A: Truncar os dados para caber no limite do modelo e avisar o usuario que os resultados podem ser parciais.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload Lattes XML File (Priority: P1)

A researcher accesses the smartLattes web application and uploads their Lattes CV XML file exported from the CNPq Plataforma Lattes. The system accepts the file, parses its XML content, converts it to a JSON structure, and stores it in the MongoDB database. The researcher receives confirmation that their CV data was successfully imported.

**Why this priority**: This is the core functionality for the acquisition context. Without it, the system has no purpose.

**Independent Test**: Can be fully tested by uploading the example file at `/docs/8334174268306003.xml` through the web interface and verifying the data appears correctly in MongoDB.

**Acceptance Scenarios**:

1. **Given** a researcher on the upload page, **When** they select a valid Lattes XML file and submit, **Then** the system parses the XML, converts it to JSON, stores it in MongoDB, and displays a success message with the researcher's name and Lattes ID.
2. **Given** a researcher on the upload page, **When** they select a file that is not a valid Lattes XML, **Then** the system displays a clear error message indicating the file is invalid and does not store anything in the database.
3. **Given** a researcher on the upload page, **When** they upload an XML for a Lattes ID that already exists in the database, **Then** the system updates (replaces) the existing record with the new data and informs the researcher that their CV was updated.

---

### User Story 2 - View Upload Confirmation Details (Priority: P2)

After a successful upload, the researcher can see a summary of the imported data, including their full name, Lattes ID number, last update date, and a count of the main sections imported (e.g., number of published articles, conference papers, software entries).

**Why this priority**: Provides immediate feedback and confidence that the upload worked correctly, but is not strictly necessary for data ingestion.

**Independent Test**: Can be tested by uploading the example XML file and verifying the confirmation summary matches the known contents of the file.

**Acceptance Scenarios**:

1. **Given** a successful XML upload, **When** the confirmation page is displayed, **Then** it shows the researcher's full name, Lattes ID (numero-identificador), and the CV last update date (data-atualizacao).
2. **Given** a successful XML upload, **When** the confirmation page is displayed, **Then** it shows counts of main production sections (bibliographic production items, technical production items, etc.).

---

### User Story 3 - Navigate Application via Main Menu (Priority: P3)

The researcher accesses the smartLattes homepage which has a main navigation menu. The menu provides access to the Upload page (acquisition context). A placeholder menu item exists for the future Data Presentation page (presentation context), which displays a "coming soon" message.

**Why this priority**: Establishes the application's navigation structure and prepares for future expansion.

**Acceptance Scenarios**:

1. **Given** a user on the homepage, **When** they look at the main navigation, **Then** they see menu items for "Enviar CV" (acquisition), "Gerar Resumo" (transformation) and "Explorar Dados" (presentation).
2. **Given** a user on the homepage, **When** they click "Enviar CV", **Then** they are taken to the XML upload page.
3. **Given** a user on the homepage, **When** they click "Gerar Resumo", **Then** they are taken to the dedicated summary generation page where they can search for a previously uploaded CV.
4. **Given** a user on the homepage, **When** they click "Explorar Dados", **Then** they see a placeholder page indicating this feature is coming soon.

---

### User Story 4 - Gerar Resumo do Pesquisador via IA (Priority: P1)

Apos o upload bem-sucedido do XML, o sistema apresenta ao pesquisador a opcao de gerar um resumo inteligente do seu perfil. O pesquisador seleciona o provedor de IA desejado (Gemini, OpenAI ou Anthropic) em um pulldown, informa a chave de API correspondente e solicita a geracao. O sistema conecta ao provedor usando a chave fornecida, seleciona o melhor modelo disponivel para aquela chave, envia os dados do curriculo junto com o prompt de geracao (armazenado em `resumoPrompt.md`) e apresenta o resumo gerado.

**Why this priority**: Este e o primeiro componente do contexto de Transformacao, que agrega valor analitico aos dados brutos do Lattes e diferencia o smartLattes de um simples repositorio de dados.

**Independent Test**: Pode ser testado fazendo upload de um XML, selecionando um provedor, informando uma chave de API valida e verificando que o resumo gerado contem analise coerente com os dados do pesquisador.

**Acceptance Scenarios**:

1. **Given** um upload bem-sucedido, **When** a confirmacao e exibida, **Then** o sistema apresenta um formulario com pulldown de provedor (Gemini, OpenAI, Anthropic) e campo para chave de API.
2. **Given** o pesquisador selecionou um provedor e informou a chave, **When** o sistema valida a chave, **Then** lista os modelos disponiveis em um segundo pulldown para o usuario escolher.
3. **Given** o pesquisador selecionou provedor, modelo e informou a chave, **When** clica em "Gerar Resumo", **Then** o sistema gera o resumo usando o modelo escolhido.
4. **Given** o resumo foi gerado, **When** e apresentado ao pesquisador, **Then** contem: analise das principais caracteristicas e pontos fortes, areas de atuacao (estruturadas), potencial de contribuicao para a ciencia, principais co-autores mais frequentes, e quantificacao da producao por area de atuacao.
5. **Given** o resumo e exibido, **When** o pesquisador deseja salva-lo, **Then** pode fazer download em formato .md, .docx ou .pdf.
6. **Given** o pesquisador fez download ou clicou em "Ok", **Then** o resumo e armazenado na colecao `resumos` do MongoDB, associado ao lattesID do pesquisador como chave primaria.

---

### User Story 5 - Tratamento de Erros na Geracao de Resumo (Priority: P2)

O sistema trata adequadamente erros na integracao com provedores de IA, exibindo mensagens claras ao pesquisador.

**Acceptance Scenarios**:

1. **Given** o pesquisador informou uma chave de API invalida, **When** o sistema tenta conectar ao provedor, **Then** exibe mensagem clara indicando que a chave e invalida ou sem permissao.
2. **Given** o provedor de IA esta indisponivel ou retorna erro, **When** a geracao falha, **Then** o sistema exibe mensagem de erro e permite que o pesquisador tente novamente.
3. **Given** o pesquisador nao selecionou um provedor ou nao informou a chave, **When** clica em "Gerar Resumo", **Then** o sistema exibe validacao indicando os campos obrigatorios.

---

### Edge Cases

- **File exceeds 10MB**: System rejects the upload immediately with a clear error message stating the file size limit.
- **Unexpected encoding**: System attempts to parse the file; if parsing fails due to encoding issues, it rejects the file with an error message. ISO-8859-1 (standard Lattes encoding) is explicitly supported (FR-008).
- **MongoDB unavailable**: System fails immediately with a clear error message informing the user the service is temporarily unavailable; user retries manually.
- **Empty or zero-byte file**: System rejects the file during validation (FR-005) with an error message indicating the file is empty or invalid.
- **Concurrent uploads**: Each upload is processed independently; MongoDB's upsert operation handles concurrent writes to the same numero-identificador safely (last write wins).
- **Special characters and HTML entities**: XML entities (e.g., `&#10;`) are resolved during parsing and preserved as-is in the JSON output.
- **Chave de API nao armazenada**: A chave de API do provedor de IA e usada apenas durante a requisicao e NAO e armazenada no servidor ou no banco de dados.
- **Timeout na geracao de resumo**: Se o provedor de IA demorar mais que 120 segundos para responder, o sistema cancela a requisicao e exibe mensagem de timeout.
- **Prompt editavel**: O prompt de geracao do resumo fica em `resumoPrompt.md` na raiz do repositorio, permitindo ajustes sem alterar codigo.
- **CV excede limite de tokens**: O sistema trunca os dados do CV para caber no contexto do modelo selecionado e avisa o usuario que os resultados podem ser parciais.

## Requirements *(mandatory)*

### Functional Requirements - Aquisicao

- **FR-001**: System MUST provide a web interface with a file upload form that accepts XML files
- **FR-002**: System MUST parse uploaded Lattes XML files and convert them to a JSON structure. De DADOS-GERAIS, armazenar apenas: nome-completo, orcid-id, nome-em-citacoes-bibliograficas, formacao-academica-titulacao, atuacoes-profissionais, areas-de-atuacao. Demais secoes (producao-bibliografica, producao-tecnica, outra-producao, dados-complementares) sao preservadas integralmente.
- **FR-003**: System MUST store the converted JSON data in the MongoDB database, using the Lattes numero-identificador as the unique document identifier. Todos os atributos DEVEM ser armazenados em caixa baixa.
- **FR-004**: System MUST handle the upsert case: if a CV with the same numero-identificador already exists, it MUST be replaced with the new upload
- **FR-005**: System MUST validate that the uploaded file is a valid Lattes XML before attempting to store it (check for CURRICULO-VITAE root element and NUMERO-IDENTIFICADOR attribute)
- **FR-006**: System MUST display appropriate success or error messages to the user after upload
- **FR-007**: System MUST provide a main navigation menu with links to "Enviar CV" (upload), "Gerar Resumo" (transformation) and "Explorar Dados" (presentation placeholder)
- **FR-008**: System MUST handle XML files encoded in ISO-8859-1 (the standard encoding used by Plataforma Lattes exports)
- **FR-009**: System MUST run as a single Docker container, published to ghcr.io/edalcin/
- **FR-010**: System MUST NOT require authentication — it operates as an open-access application on a trusted network
- **FR-011**: System MUST reject uploaded files larger than 10MB with a clear error message
- **FR-012**: System MUST display a clear error message when the MongoDB database is unavailable, without queuing or retrying the upload
- **FR-013**: System MUST NOT store dados pessoais ou sensiveis (CPF, RG, passaporte, raca, cor, endereco, telefone, nacionalidade, sexo, etc.) no banco de dados

### Functional Requirements - Transformacao

- **FR-100**: O sistema DEVE apresentar formulario para geracao de resumo do pesquisador com: pulldown de selecao de provedor de IA (Gemini, OpenAI, Anthropic), campo para chave de API e pulldown de modelo. Este formulario DEVE estar disponivel (a) apos upload bem-sucedido e (b) em uma pagina dedicada "Gerar Resumo" acessivel pelo menu principal, onde o usuario pode buscar um CV ja armazenado pelo lattesID ou nome.
- **FR-101**: O sistema DEVE integrar com tres provedores de IA: Google Gemini, OpenAI e Anthropic. A integracao DEVE ser feita via API REST oficial de cada provedor.
- **FR-102**: Apos o usuario selecionar o provedor e informar a chave de API, o sistema DEVE listar os modelos disponiveis para aquela chave (via API do provedor) e apresenta-los em um segundo pulldown para que o usuario escolha o modelo desejado.
- **FR-102a**: Na pagina dedicada "Gerar Resumo", o sistema DEVE fornecer um campo de busca por texto que pesquisa por lattesID ou nome do pesquisador nos CVs armazenados no MongoDB.
- **FR-103**: O sistema DEVE enviar ao provedor de IA os dados do curriculo do pesquisador (armazenados no MongoDB) junto com o prompt definido em `resumoPrompt.md`.
- **FR-104**: O resumo gerado DEVE conter: (a) analise das principais caracteristicas e pontos fortes do pesquisador, (b) areas de atuacao de forma estruturada, (c) potencial de contribuicao para a ciencia, (d) principais e mais frequentes co-autores, (e) quantificacao da producao por area de atuacao.
- **FR-105**: O sistema DEVE apresentar o resumo gerado ao usuario com opcoes de download em .md, .docx e .pdf.
- **FR-106**: O sistema DEVE armazenar o resumo gerado na colecao `resumos` do MongoDB, usando o lattesID como chave primaria (`_id`).
- **FR-107**: A chave de API do provedor de IA NAO DEVE ser armazenada no servidor ou banco de dados. Deve ser usada apenas durante a requisicao de geracao.
- **FR-108**: O prompt de geracao do resumo DEVE ser lido do arquivo `resumoPrompt.md` na raiz do repositorio, permitindo edicao sem alterar codigo.
- **FR-109**: O sistema DEVE exibir mensagens de erro claras para: chave de API invalida, provedor indisponivel, timeout (120s) e demais falhas de integracao.
- **FR-110**: Quando os dados do CV excederem o limite de contexto do modelo selecionado, o sistema DEVE truncar os dados para caber no limite e exibir aviso ao usuario informando que os resultados podem ser parciais.

### Key Entities

- **Curriculo (CV)**: Representa o curriculo Lattes de um pesquisador. Identificado por `numero-identificador` (chave primaria). De DADOS-GERAIS, armazena apenas: nome-completo, orcid-id, nome-em-citacoes-bibliograficas, formacao-academica-titulacao, atuacoes-profissionais, areas-de-atuacao. Demais secoes (producao-bibliografica, producao-tecnica, outra-producao, dados-complementares) sao preservadas integralmente. Todos os atributos em caixa baixa.
- **Upload Event**: Represents a single file upload action. Contains the original filename, upload timestamp, processing status, and the extracted numero-identificador.
- **Resumo do Pesquisador**: Artefato gerado por IA a partir dos dados do curriculo. Armazenado na colecao `resumos` do MongoDB com `_id` = lattesID. Contem: analise de pontos fortes, areas de atuacao estruturadas, potencial cientifico, co-autores frequentes e producao quantificada por area. Inclui metadados: provedor e modelo de IA utilizados, data de geracao.

## Constraints & Assumptions

- **C-001**: The MongoDB database runs at `192.168.1.10:27017` with database name `smartLattes`
- **C-002**: The application MUST be packaged as a single Docker container
- **C-003**: The Docker image MUST be published to `ghcr.io/edalcin/`
- **C-004**: The Docker image should be as small as possible
- **C-005**: The technology stack should be simple and modern, enabling a clean web interface
- **C-006**: The application follows the C4 Model architecture with three contexts: acquisition (upload), transformation (AI-powered analysis), and presentation (future data visualization)
- **C-007**: An example Lattes XML file is available at `/docs/8334174268306003.xml` for reference and testing
- **C-008**: As chaves de API dos provedores de IA sao fornecidas pelo usuario a cada uso e nao sao persistidas
- **C-009**: O prompt de geracao e mantido em `resumoPrompt.md` no repositorio, versionado junto com o codigo
- **C-010**: A comunicacao com provedores de IA e feita server-side (backend Go) para proteger a chave de API do usuario
- **C-011**: A aplicacao serve via HTTP puro (sem TLS). A chave de API transita em texto claro na rede interna confiavel. Se necessario, um reverse proxy externo pode ser adicionado para TLS.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A researcher can upload a Lattes XML file and see a success confirmation within 10 seconds for files up to 1MB
- **SC-002**: The uploaded data is correctly stored in MongoDB and can be queried by numero-identificador
- **SC-003**: The system correctly handles the example file at `/docs/8334174268306003.xml`, preserving the sections defined in FR-002
- **SC-004**: Invalid files (non-XML, non-Lattes XML) are rejected with a user-friendly error message
- **SC-005**: The Docker image size is under 200MB
- **SC-006**: The application starts and is ready to accept uploads within 10 seconds of container start
- **SC-007**: Re-uploading a CV with the same numero-identificador updates the existing record without creating duplicates
- **SC-008**: O pesquisador consegue gerar um resumo via qualquer um dos tres provedores (Gemini, OpenAI, Anthropic) com uma chave de API valida
- **SC-009**: O resumo gerado contem todos os cinco elementos definidos em FR-104
- **SC-010**: O resumo e armazenado corretamente na colecao `resumos` do MongoDB, associado ao lattesID
- **SC-011**: O pesquisador consegue fazer download do resumo nos tres formatos (.md, .docx, .pdf)
- **SC-012**: Nenhum dado pessoal ou sensivel (CPF, RG, endereco, etc.) e armazenado no banco de dados
