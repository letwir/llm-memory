# Knowledge: llm-memory Issues follow-up

- task_id: 2026-08-29-llm-memory-issues-followup
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace `A:\Users\letwir\repo\llm-memory`, PostgreSQL 18 remote database
- evidence/source scope: `Issues.md`, Go source, `init.sql`, user-reported PostgreSQL migration and verification output
- verification status: local Go tests pass; database migration and index creation were confirmed by user output
- remaining uncertainty: embedding検索、実機競合テスト、WORM境界は未検証

The current search baseline is PostgreSQL trigger-maintained `tsvector` plus existing substring search. The active identity index requires category-specific normalized titles, so source/task-specific titles are required by operation.
