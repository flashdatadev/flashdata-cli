# Search monitoring recipe

Use this to capture or compare search-result snapshots for a topic. FlashData provides query results; this Skill is not a scheduler and does not monitor arbitrary page bodies.

## Capture comparable snapshots

Preserve the query, requested result count, page, region, language, and filters across runs. `google_search` uses `num` for requested results and `page` for pagination; every additional page is another paid query. Region/language defaults come from the provider, so set them explicitly when comparability matters.

For an initial run, save the observed results as a baseline at the user's chosen local destination. Include retrieval time, query settings, source URLs, titles, snippets, and ranks only when returned. Report it as a baseline, not as a discovered change. If no previous snapshot is available, do not fabricate one.

For later runs, match by source URL using only defensible normalization, such as removing a fragment. Preserve query parameters unless you know they are tracking-only; they can identify different resources. Compare actual returned ranks and content. Label new, missing, and changed entries as changes in the observed search results, not proof that a page was created, deleted, or edited.

Search rankings and coverage fluctuate. A changed snippet is not full-page content verification. Distinguish provider publication dates from the time you retrieved the snapshot.

## Recurrence and delivery

Use an available scheduler only when the user asks for recurring execution. Establish the schedule/timezone, per-run query bound, destination, and notification preference as needed. A request to compare two snapshots does not create a schedule. Do not claim an automation exists until the actual scheduler confirms it.

For recurring work, keep failed or incomplete runs separate from the last successful baseline; a transient failure must not mark every source as removed. On uncertain submission outcomes, use the core Skill's job-recovery guidance rather than repeating paid searches. Send reports to an external service only within the user's authorized destination and scope.
