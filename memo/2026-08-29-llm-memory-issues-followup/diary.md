# Diary: llm-memory Issues follow-up

- timestamp: 2026-08-29 Asia/Tokyo
- task: continue ordered llm-memory issue resolution and close out only when no issues remain
- request-evidence: user requested ingesting Issues.md and deleting it only if no issues remain, then committing
- action: applied the ordered implementation batch, corrected the invalid generated-column migration to a trigger migration, and verified user DB results
- result: local tests pass; Issues.md remains because three substantive items are still open
- friction: generated-column immutability and title-only uniqueness were invalid assumptions and were corrected
- attribution: implementation corrections were agent-owned; user request was sufficiently scoped
- impact: no destructive deletion performed because the stated condition was false
- feedback: keep Issues.md until embedding and production-boundary verification are complete
- rewritten-request: ingest the current Issues.md and task records, retain Issues.md while substantive issues remain, and commit the verified workspace changes
