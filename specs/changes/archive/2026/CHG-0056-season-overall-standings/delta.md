# CHG-0056 - Season-wide standings behavioral delta

## US-0056-001 - Compare my season-wide progress

### REQ-0056-001 - Rank competitors within one season

ADDED (canonical target REQ-SCC-020): sum merits across all categories
for each distinct recipient on the INDIVIDUAL track in one season.
Rank by total descending, then distinct completed contract count,
then distinct category/objective pairs, then recipient ID as the
deterministic final tie-break. A competitor with credits in multiple
categories appears once; TEAM credits and other seasons never contribute.
Participants are distinct credited recipients, not the sum of category
participant counts. Existing category rankings are unchanged.

#### SCN-0056-001 - Reward progress across categories

Given Mira has 120 merits in air combat and 80 in racing while Kai
has 190 in air combat in the same season
When a member views the overall individual standings
Then Mira leads with 200 merits and Kai follows with 190.

#### SCN-0056-002 - Keep tracks and seasons separate

Given Mira also has team merits and merits from an earlier season
When a member views this season's individual standings
Then those merits do not change Mira's overall position.

#### SCN-0056-003 - Resolve ties predictably

Given two competitors have equal merits
When their overall positions are determined
Then distinct completed contracts, category/objective pairs, and
finally stable recipient identity determine their order.

### REQ-0056-002 - Show my honest position

ADDED (canonical target REQ-SCC-021): the same overall result contains
the authenticated member's position, derived from their session identity,
even when they are outside the requested top page. A member without
credited merits has no rank, not rank zero. No unrelated identity,
evidence, contracts, or private metadata is returned.

#### SCN-0056-004 - Show a position outside the top page

Given Kai is ranked 23rd and the top page shows ten members
When Kai views the overall standings
Then Kai sees their true position 23 alongside the top ten.

#### SCN-0056-005 - Do not invent a first position

Given a new season has no credited merits
When a member views the overall standings
Then the page is empty and their position is absent.

#### SCN-0056-006 - Preserve safe access and navigation

Given a season has multiple pages of ranked competitors
When an authenticated member requests subsequent pages
Then rows retain their deterministic order without duplicates;
an unauthenticated member sees no standings.

## Compatibility

ADDED: one new read-only endpoint. MODIFIED: no existing API behavior.
REMOVED: nothing; existing per-category and history readers remain
actively used. A future removal requires an approved delta and a
consumer/compatibility inventory, not a speculative cleanup.
