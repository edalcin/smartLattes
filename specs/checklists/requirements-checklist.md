# Requirements Checklist: Lattes XML Upload

**Purpose**: Track verification of all functional requirements and success criteria for the Lattes XML Upload feature
**Created**: 2026-02-07
**Feature**: [spec.md](../spec.md)

## Functional Requirements

- [ ] CHK001 FR-001: Web interface provides a file upload form that accepts XML files
- [ ] CHK002 FR-002: System parses Lattes XML and converts to JSON preserving all 5 top-level sections (DADOS-GERAIS, PRODUCAO-BIBLIOGRAFICA, PRODUCAO-TECNICA, OUTRA-PRODUCAO, DADOS-COMPLEMENTARES)
- [ ] CHK003 FR-003: Converted JSON is stored in MongoDB using NUMERO-IDENTIFICADOR as unique document identifier
- [ ] CHK004 FR-004: Uploading a CV with an existing NUMERO-IDENTIFICADOR replaces the previous record (upsert)
- [ ] CHK005 FR-005: System validates uploaded file is a valid Lattes XML (checks for CURRICULO-VITAE root element and NUMERO-IDENTIFICADOR attribute)
- [ ] CHK006 FR-006: Success and error messages are displayed to the user after upload attempt
- [ ] CHK007 FR-007: Main navigation menu exists with links to Upload page and placeholder for Data Presentation
- [ ] CHK008 FR-008: System correctly handles XML files encoded in ISO-8859-1
- [ ] CHK009 FR-009: Application runs as a single Docker container published to ghcr.io/edalcin/
- [ ] CHK035 FR-010: System does not require authentication (open access)
- [ ] CHK036 FR-011: Files larger than 10MB are rejected with clear error message
- [ ] CHK037 FR-012: System shows clear error when MongoDB is unavailable (no queue/retry)

## User Story Acceptance

### US1 - Upload Lattes XML File

- [ ] CHK010 Valid Lattes XML upload results in successful parse, JSON conversion, MongoDB storage, and success message with researcher name and Lattes ID
- [ ] CHK011 Invalid file upload (non-XML or non-Lattes XML) shows clear error message and stores nothing in database
- [ ] CHK012 Re-upload of same Lattes ID updates existing record and informs user of update

### US2 - View Upload Confirmation Details

- [ ] CHK013 Confirmation page shows researcher full name, Lattes ID, and CV last update date
- [ ] CHK014 Confirmation page shows counts of main production sections

### US3 - Navigate Application via Main Menu

- [ ] CHK015 Homepage displays main navigation menu with "Enviar CV" and "Explorar Dados" items
- [ ] CHK016 "Enviar CV" menu item navigates to the upload page
- [ ] CHK017 "Explorar Dados" menu item shows a "coming soon" placeholder page

## Edge Cases

- [ ] CHK018 Files exceeding 10MB size limit are rejected with appropriate message
- [ ] CHK019 Empty or zero-byte files are rejected with appropriate message
- [ ] CHK020 System handles XML with special characters and HTML entities (e.g., `&#10;`)
- [ ] CHK021 System displays appropriate error when MongoDB connection is unavailable

## Success Criteria

- [ ] CHK022 SC-001: Upload completes within 10 seconds for files up to 1MB
- [ ] CHK023 SC-002: Uploaded data is queryable in MongoDB by NUMERO-IDENTIFICADOR
- [ ] CHK024 SC-003: Example file (/docs/8334174268306003.xml) is processed correctly with all 5 sections preserved
- [ ] CHK025 SC-004: Invalid files are rejected with user-friendly error messages
- [ ] CHK026 SC-005: Docker image size is under 200MB
- [ ] CHK027 SC-006: Application starts and is ready within 10 seconds of container start
- [ ] CHK028 SC-007: No duplicate records created when same CV is uploaded multiple times

## Constraints Verification

- [ ] CHK029 C-001: Application connects to MongoDB at 192.168.1.10:27017 database smartLattes
- [ ] CHK030 C-002: Application packaged as single Docker container
- [ ] CHK031 C-003: Docker image published to ghcr.io/edalcin/
- [ ] CHK032 C-004: Docker image uses minimal base image for smallest possible size
- [ ] CHK033 C-005: Technology stack is simple and modern
- [ ] CHK034 C-006: Architecture follows C4 Model with acquisition and presentation contexts

## Notes

- Check items off as completed: `[x]`
- Add comments or findings inline
- The example XML file at `/docs/8334174268306003.xml` (~304KB, ISO-8859-1 encoded) should be used as the primary test artifact
- Items are numbered sequentially (CHK001-CHK037) for easy reference
