# Dashboard

`GET /v1/admin/dashboard?days=7` (or `days=30`) requires an administrator Bearer token. This endpoint queries PostgreSQL on each request and returns an uncached snapshot. No new migration is needed.

- `totals`: all stored users, system videos, channels, original caption sentences for system videos, missing captions, active learnable languages, active topics, and levels. `activeUsers` means enabled accounts, not online users.
- `period`: registrations, new system videos, and distinct users whose latest watch history falls inside the selected period.
- `daily`: exactly 7 or 30 calendar days including today, with zero-filled dates, in `Asia/Ho_Chi_Minh`. The legacy `users.created_at` column stores UTC without a timezone.
- `recentVideos`: five newest system videos, ordered by creation time and ID.

User imports are excluded from system video/caption metrics and returned separately as `userImportedVideos`.

Run the read-only PostgreSQL smoke test against an existing migrated database by setting `DASHBOARD_TEST_DATABASE_URL`, then running:

```powershell
go test -count=1 -v ./internal/repo/persistent/dashboard
```

For exact count assertions, use a separate empty test database, apply each project migration in a transaction, apply `testdata/fixture.sql`, and set `DASHBOARD_TEST_FIXTURE=1`. This checks private imports, disabled accounts/channels, caption counts, distinct learners, recent ordering, and Vietnam midnight boundaries. Never load this fixture into application data.
