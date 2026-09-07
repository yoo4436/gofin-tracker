---
name: gofin-db-migration
description: Inspect, author, and locally validate GoFin-Tracker Supabase schema migrations, including tables, functions, RLS, grants, and publication logic. Explanation and inventory alone do not authorize mutation.
---

# GoFin database migration

Read repository AGENTS.md and establish the task's authorization.

## Inspect before changing

- Read relevant migrations, application consumers and configuration without printing secrets. Identify the local versus remote target explicitly.
- Inspect migration history when relevant and accessible; report unavailable access rather than assuming histories match.
- Explain schema, data and permission effects. `db pull` retrieves schema rather than remote data; `migration repair` changes tracking rather than applying SQL.

## Local implementation and verification

- Add a migration for shared schema changes; do not rewrite applied history to conceal discrepancies.
- Preserve affected policies, grants, constraints, indexes and application contracts. For daily report publication, inspect `public.publish_daily_report` and preserve its validation rather than setting publication status directly.
- Check local Supabase and Docker availability. A reset destroys local database data: confirm the target and existing authorization for disposable data/reset, or request that specific authorization before resetting.
- When authorized, run `supabase db reset --local`, `supabase db lint --local` and `supabase db diff --local`. Inspect warnings and residual differences against the known baseline.
- Run affected application tests when behavior changes. Report unexecuted checks and unavailable infrastructure honestly.

## Remote boundary

- Do not run remote reset, `supabase db push`, migration repair or remote schema/data writes without explicit authorization for the action and target.
- Complete the reviewable migration and available local checks before requesting remote application approval; do not repeat existing authorization requests.
- Report paths, compatibility/data effects, actual checks, unresolved differences and whether any remote change occurred.
