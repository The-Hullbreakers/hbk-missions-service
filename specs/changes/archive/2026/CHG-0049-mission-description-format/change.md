---
change_id: CHG-0049
title: Format mission descriptions as markdown with structured rules
status: completed
risk: L2
canonical_action: modify
canonical_paths:
  - specs/capabilities/mission-board/spec.md
execution_mode: sequential
primary_workflow: WORKFLOW-GO_SDD_IMPLEMENT_CHANGE_WORKFLOW
supporting_workflows:
  - WORKFLOW-COMMON_BDD_SPECIFICATION_WORKFLOW
  - WORKFLOW-COMMON_SDD_CHANGE_LIFECYCLE_WORKFLOW
---

# Mission description format

User request: mission descriptions arrive at the frontend as one unreadable
paragraph because seeds concatenate every rule sentence into a single string.
Approved scope (Gate 1): two complementary changes —

- A: `Description` is declared Markdown content so clients can render
  paragraphs, emphasis, and lists.
- B: `Rules` is a new structured list of short requirement statements on
  definitions and offers so clients render bullets without parsing prose.

## Acceptance scenarios

- A definition created with markdown `description` and a `rules` list echoes
  both verbatim; `rules` items are non-empty strings, no minimum count.
- An offer published from a version freezes `description` (markdown) and
  `rules` alongside title/objectives; republishing a new version does not
  alter live offers or contracts.
- `GET /api/v1/missions/offers`, `GET /offers/{id}`, `POST /offers`,
  `GET /definitions` and contract detail expose `Rules` plus the unchanged
  `Description` string; legacy stored items without the attribute decode an
  empty list.
- Seeds emit a markdown `Description` (short intro paragraph) and move the
  shared requirement sentences into `Rules` items.

## Layer gates / verification

Domain: `MissionDefinition`/`MissionOffer` carry `Rules`; freeze in
`NewMissionOfferFromDefinition`. Application: create-definition passes
`rules`. Boundary: request/response projections. Infrastructure: DynamoDB
list attribute with legacy-tolerant decode. Seeds: script + catalogs emit
markdown + rules. Tests-first per wave in `tasks.md`.

Rollback: additive field and content-format only; republish a corrected
definition version to amend content. Legacy items degrade to `rules: []`.
