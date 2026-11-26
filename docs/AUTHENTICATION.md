# Authentication And Authorization

## Role

Missions is an OAuth 2.0 resource server. Identity issues and validates the origin of the identity; Missions applies contextual authorization over its resources.
The effective caller comes only from the validated token. `player_id`, `roles` and `team_ids` are never accepted from body, query, path or alternative header.
## Business Capabilities

| Capability | Allows |
| --- | --- |
| `missions:offer:read` | Consult the mission board. |
| `missions:contract:accept` | Accept your own contract. |
| `missions:contract:participate` | Progress or deliver their own or team contract. |
| `missions:definition:manage` | Design and publish missions. |
| `missions:adjudication:review` | Accept/reject evidence and resolve disputes. |
| `missions:season:manage` | Set up, freeze and end season. |
| `missions:audit:read` | Consult evidence and settings. |
Capability does not replace ownership: one participant does not modify another's contract; a reviewer does not award points; a manager does not edit final standings.
## Single Scope

- The platform runs one global scope: definitions, offers, contracts, limits, seasons and rankings are shared by every caller.
- Reads address a resource directly by ID; there is no tenancy prefix or isolation boundary.
- The user cannot choose recipient or team from non-validated data.
## Team Authorization

Party Manager is the authority of the live party. Missions obtains a snapshot, demands confirmation, and then has only the frozen contractual roster. Subsequent party changes do not rewrite the contract.
## Competitive Integrity

Only authorized adjudications create confirmations. Merits, category, season and track are derived from published rules and local status. Adjustments require a separate reason and authorization; all administrative action is auditable.
