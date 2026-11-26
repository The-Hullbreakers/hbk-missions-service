# CHG-0054 — Catalog translation drafts

Current task: **T-0054-014 — Draft English for all catalog missions**
(in progress, 4/36). Next action within T-0054-014: author the four
PvP air definitions under `seeds/translations/`, followed by the four
PvP duel definitions and 24 racing definitions. Keep each new JSON
file below 150 physical lines. Do not mark text reviewed; operator
semantic approval is required separately.

Gate 1 approved for catalog translation artifacts and safe redeploy.
Gate 2 approved only for T-0054-013 unit RED and T-0054-014 Git drafts.
Gate 2 for T-0054-015 DynamoDB seed RED and Gate 3-SEED for any script
code are **not approved**. No real backfill, AWS query, table writes,
or deployment authorized. Preserve CHG-0051 and unrelated work.

TEST-SEED-LANG-001 lives in `tests/unit/seeds/test_mission_translations.py`
with fixture `translation_fixture.py` (90 and 60 lines). Command:
`python3 -m unittest discover -s tests/unit/seeds -p 'test_mission_translations.py' -v`.
Actual RED 2026-11-02: three top-level tests, three errors and seven
failures from the absent `seed.translated_definitions` function. All
test fixtures imported; no database or network used. This is an
API-absence RED, not yet runtime proof of validator behavior.

One sidecar exists: `seeds/translations/pvp-fps-v4.en.json` with four
FPS records (116 lines). Their `SourceSpanish` and version match
`seed.definitions(..., "staging")` byte-for-byte for text and IDs;
English rule count and objective IDs match. The originals in
`seeds/pvp-missions.json` are unchanged. `seeds/translations/README.md`
labels drafts explicitly unreviewed. No seed-script production changes.

Design contract: each sidecar JSON has `translations` entries with
`DefinitionCode`, `DefinitionVersion`, `SourceSpanish`, `English`;
both locales contain `Title`, `Description`, ordered `Rules`, and
`Objectives` (`ObjectiveID`, `Name`, `Description`). Spanish stays
the canonical source, no economics in translations, no automated
semantic approval. `design.md`, `delta.md`, `tasks.md`,
`traceability.yaml`, `verification.md`, and `rollout.md` map this work.
The CI racing seed runs after deploy and currently rejects or erases
same-version English; do not attempt a real seed until Gate 3-SEED
and the redeploy test are GREEN.

Next work: translate PvP air/duel and racing, verifying each sidecar
source/version/structure against the Spanish seed after writing; do
not use the absent API as an expectation of GREEN. Racing has six
unique Spanish rule strings (five universal, one staging-only),
one objective type, three track names, and four difficulty names;
PvP has 12 definitions, three families with shared rules. Avoid
changing Spanish grammar or gameplay/economics while translating.
After all 36 drafts, record result; obtain Gate 2 for T-0054-015 RED
on safe seed and ask Gate 3-SEED only after reviewing its actual RED.
