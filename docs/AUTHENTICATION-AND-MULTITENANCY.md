# Authentication, Authorization And Multi-Tenancy

## Role

Missions is an OAuth 2.0 resource server. Identity issues and validates the origin of the identity; Missions applies contextual authorization over its resources.
The effective tenant comes only from the validated token. `tenant_id` is never accepted from body, query, path or alternative header.
## Business Capabilities

| Capability | Allows |
| --- | --- |
| `missions:offer:read` | Consult the tenant board. |
| `missions:contract:accept` | Accept your own contract. |
| `missions:contract:participate` | Progress or deliver their own or team contract. |
| `missions:definition:manage` | Design and publish missions. |
| `missions:adjudication:review` | Accept/reject evidence and resolve disputes. |
| `missions:season:manage` | Set up, freeze and end season. |
| `missions:audit:read` | Consult evidence and tenant settings. |
Capability does not replace ownership: one participant does not modify another's contract; a reviewer does not award points; a manager does not edit final standings.
## Tenant Isolation

- Pool, definitions, contracts, limits, seasons and rankings are tenant-scoped.
- All reading starts with tenant before the resource ID.
- An ID from another tenant responds as non-existent in public boundaries.
- Caches, deduplication and events include tenant.
- The user cannot choose recipient or team from non-validated data.
## Team Authorization

Party Manager is the authority of the live party. Missions obtains a snapshot, demands confirmation, and then has only the frozen contractual roster. Subsequent party changes do not rewrite the contract.
## Competitive Integrity

Only authorized adjudications create confirmations. Merits, category, season and track are derived from published rules and local status. Adjustments require separate reason and capacity; all administrative action is auditable.
