# Feature Specification: Lattes XML Upload

**Feature**: `lattes-xml-upload`
**Created**: 2026-02-07
**Status**: Draft
**Input**: User description: "Este projeto vai alimentar uma base de dados MongoDB com conteudo exportado em XML da Plataforma Lattes, do CNPq. O banco de dados roda em 192.168.1.10:27017 e tem o nome de 'smartLattes'. O projeto tera uma interface web onde o usuario fara upload do arquivo XML que representa seu perfil de pesquisador na plataforma Lattes. Um arquivo exemplo esta em /docs/8334174268306003.xml. Este arquivo enviado pela interface web sera convertido para uma estrutura em JSON e carregado no banco de dados MongoDB. Inicialmente a interface web tera apenas esta funcionalidade de upload do arquivo XML (contexto de aquisicao, segundo C4 Model). Porem, no futuro tera uma interface adicional, acessada via menu principal na homepage, que sera de apresentacao dos dados na base de dados (contexto de apresentacao, segundo C4 Model). Este projeto devera rodar completamente em apenas um docker, que sera gerado e ficara exposto em ghcr.io/edalcin/. Procure um stack tecnologico simples mas moderno, que permita a criacao de uma interface simples e moderna, e que gere um arquivo docker no menor tamanho possivel."

## Clarifications

### Session 2026-02-07

- Q: Does the system require authentication or access control? → A: No authentication — open access, intended for use on a trusted/internal network.
- Q: What is the maximum upload file size? → A: 10MB maximum per uploaded XML file.
- Q: What should happen when MongoDB is unavailable at upload time? → A: Fail immediately with a clear error message; the user retries manually.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Upload Lattes XML File (Priority: P1)

A researcher accesses the smartLattes web application and uploads their Lattes CV XML file exported from the CNPq Plataforma Lattes. The system accepts the file, parses its XML content, converts it to a JSON structure, and stores it in the MongoDB database. The researcher receives confirmation that their CV data was successfully imported.

**Why this priority**: This is the core and only functionality for the initial release (acquisition context). Without it, the system has no purpose. It delivers the primary value of ingesting Lattes data into the database.

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

1. **Given** a successful XML upload, **When** the confirmation page is displayed, **Then** it shows the researcher's full name, Lattes ID (NUMERO-IDENTIFICADOR), and the CV last update date (DATA-ATUALIZACAO).
2. **Given** a successful XML upload, **When** the confirmation page is displayed, **Then** it shows counts of main production sections (bibliographic production items, technical production items, etc.).

---

### User Story 3 - Navigate Application via Main Menu (Priority: P3)

The researcher accesses the smartLattes homepage which has a main navigation menu. The menu provides access to the Upload page (acquisition context). A placeholder menu item exists for the future Data Presentation page (presentation context), which displays a "coming soon" message.

**Why this priority**: Establishes the application's navigation structure and prepares for future expansion, but the upload functionality can work without a formal menu in the MVP.

**Independent Test**: Can be tested by navigating to the homepage, verifying the menu is visible, clicking the Upload link and reaching the upload page, and clicking the Presentation link and seeing the placeholder message.

**Acceptance Scenarios**:

1. **Given** a user on the homepage, **When** they look at the main navigation, **Then** they see menu items for "Enviar CV" (acquisition) and "Explorar Dados" (presentation).
2. **Given** a user on the homepage, **When** they click "Enviar CV", **Then** they are taken to the XML upload page.
3. **Given** a user on the homepage, **When** they click "Explorar Dados", **Then** they see a placeholder page indicating this feature is coming soon.

---

### Edge Cases

- **File exceeds 10MB**: System rejects the upload immediately with a clear error message stating the file size limit.
- **Unexpected encoding**: System attempts to parse the file; if parsing fails due to encoding issues, it rejects the file with an error message. ISO-8859-1 (standard Lattes encoding) is explicitly supported (FR-008).
- **MongoDB unavailable**: System fails immediately with a clear error message informing the user the service is temporarily unavailable; user retries manually.
- **Empty or zero-byte file**: System rejects the file during validation (FR-005) with an error message indicating the file is empty or invalid.
- **Concurrent uploads**: Each upload is processed independently; MongoDB's upsert operation handles concurrent writes to the same NUMERO-IDENTIFICADOR safely (last write wins).
- **Special characters and HTML entities**: XML entities (e.g., `&#10;`) are resolved during parsing and preserved as-is in the JSON output.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a web interface with a file upload form that accepts XML files
- **FR-002**: System MUST parse uploaded Lattes XML files and convert them to a JSON structure preserving all data sections (DADOS-GERAIS, PRODUCAO-BIBLIOGRAFICA, PRODUCAO-TECNICA, OUTRA-PRODUCAO, DADOS-COMPLEMENTARES)
- **FR-003**: System MUST store the converted JSON data in the MongoDB database, using the Lattes NUMERO-IDENTIFICADOR as the unique document identifier
- **FR-004**: System MUST handle the upsert case: if a CV with the same NUMERO-IDENTIFICADOR already exists, it MUST be replaced with the new upload
- **FR-005**: System MUST validate that the uploaded file is a valid Lattes XML before attempting to store it (check for CURRICULO-VITAE root element and NUMERO-IDENTIFICADOR attribute)
- **FR-006**: System MUST display appropriate success or error messages to the user after upload
- **FR-007**: System MUST provide a main navigation menu with links to the Upload page and a placeholder for the future Data Presentation page
- **FR-008**: System MUST handle XML files encoded in ISO-8859-1 (the standard encoding used by Plataforma Lattes exports)
- **FR-009**: System MUST run as a single Docker container, published to ghcr.io/edalcin/
- **FR-010**: System MUST NOT require authentication — it operates as an open-access application on a trusted network
- **FR-011**: System MUST reject uploaded files larger than 10MB with a clear error message
- **FR-012**: System MUST display a clear error message when the MongoDB database is unavailable, without queuing or retrying the upload

### Key Entities

- **Curriculo (CV)**: Represents a researcher's complete Lattes CV. Identified uniquely by `NUMERO-IDENTIFICADOR`. Contains sections: general data (DADOS-GERAIS with personal info, education, professional activities, areas of expertise, languages), bibliographic production (PRODUCAO-BIBLIOGRAFICA with conference papers, published articles, books/chapters, newspaper/magazine texts), technical production (PRODUCAO-TECNICA with software, technical work), other production (OUTRA-PRODUCAO), and complementary data (DADOS-COMPLEMENTARES). Root element attributes include system origin, ID number, and last update date/time.
- **Upload Event**: Represents a single file upload action. Contains the original filename, upload timestamp, processing status, and the extracted NUMERO-IDENTIFICADOR. Used for tracking and displaying confirmation.

## Constraints & Assumptions

- **C-001**: The MongoDB database runs at `192.168.1.10:27017` with database name `smartLattes`
- **C-002**: The application MUST be packaged as a single Docker container
- **C-003**: The Docker image MUST be published to `ghcr.io/edalcin/`
- **C-004**: The Docker image should be as small as possible
- **C-005**: The technology stack should be simple and modern, enabling a clean web interface
- **C-006**: The application follows the C4 Model architecture with two contexts: acquisition (this feature) and presentation (future feature)
- **C-007**: An example Lattes XML file is available at `/docs/8334174268306003.xml` for reference and testing

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A researcher can upload a Lattes XML file and see a success confirmation within 10 seconds for files up to 1MB
- **SC-002**: The uploaded data is correctly stored in MongoDB and can be queried by NUMERO-IDENTIFICADOR
- **SC-003**: The system correctly handles the example file at `/docs/8334174268306003.xml`, preserving all 5 top-level sections (DADOS-GERAIS, PRODUCAO-BIBLIOGRAFICA, PRODUCAO-TECNICA, OUTRA-PRODUCAO, DADOS-COMPLEMENTARES)
- **SC-004**: Invalid files (non-XML, non-Lattes XML) are rejected with a user-friendly error message
- **SC-005**: The Docker image size is under 200MB
- **SC-006**: The application starts and is ready to accept uploads within 10 seconds of container start
- **SC-007**: Re-uploading a CV with the same NUMERO-IDENTIFICADOR updates the existing record without creating duplicates
