# NES-006 Template

## 1. Status
- Status: Draft
- Version: 0.1
- Owner: NAEOS Core Team

## 2. Purpose
This specification defines reusable templates used to produce project artifacts in a consistent format.

## 3. Scope
The template framework covers document templates, source templates, pipeline templates, infrastructure templates, and policy templates.

## 4. Requirements
### 4.1 Functional Requirements
- FR-001: The template system shall support parameterization.
- FR-002: The template system shall preserve the required structure for downstream generation.

### 4.2 Non-Functional Requirements
- NFR-001: Templates shall be versioned and maintainable.
- NFR-002: Templates shall remain reusable across multiple projects.

## 5. Template Categories
- Document templates
- Source code templates
- CI/CD templates
- Infrastructure templates
- Governance templates

## 6. Workflow
1. Select the appropriate template.
2. Supply the required parameters.
3. Generate the target artifact.
4. Validate the generated result.

## 7. Acceptance Criteria
- A template can be instantiated repeatedly with consistent output structure.
- Generated artifacts conform to the expected template contract.
