---
change_id: CHG-0054
artifact: mutation-report
status: completed
updated: 2026-11-16
---

# Mutation Report — CHG-0054

## Scope

Non-trivial logic selected for mutation: `lang` query resolution and
fail-closed validation, served-language labeling, Spanish fallback under
`lang=en`, English-snapshot completeness on publish, frozen-snapshot
immutability on acceptance, and conditional source-checked backfill.

## Method

`go-mutesting`/`gremlins` are not installed; manual analysis of critical
mutations, matching the convention recorded in CHG-0003.

## Critical Mutations Reviewed

| Function | Mutation | Killing test |
| --- | --- | --- |
| `requestedMissionLanguage` | Default `en` instead of `es`, or accept `fr` | `localized_*` HTTP suites fail — omitted `lang` must serve `es`; unsupported/conflicting values must return `400 INVALID_QUERY_PARAMETER` |
| served-language label | Emit `en` while serving Spanish fallback | `localized_list_offers*_test` / detail tests fail — items must declare the language actually served |
| `Content-Language` header | Always emit, or emit on mixed collections | `localized_list_*_mixed_test` fail — header only on complete homogeneous payloads |
| Spanish fallback | Reject `lang=en` on es-only records (406/404) | localized offer/detail tests fail — legacy records serve `es` labeled `es` |
| `englishSnapshotComplete` | Skip objective/rule coverage check | `mission_definition_locales_*_test` fail — partial snapshots must block publish |
| Offer/contract snapshot freeze | Reference live definition instead of cloned snapshot | TEST-E2E-LANG-001 fails — post-accept translation edits must not reach the contract |
| Backfill conditional write | Unconditional `PutItem` or es mutation | `localized_backfill_*_test` fail — drift conflicts abort, Spanish rows untouched |

## Conclusion

- Every scoped mutant is killed by an existing unit/integration/composition
  test or the executed critical E2E.
- Translation-content correctness is source-checked by the reviewed
  translation suite and backfill rehearsal, not mutatable code.
- Residual: qualitative manual record, not a tool-computed score — same
  declared limitation as CHG-0003.
