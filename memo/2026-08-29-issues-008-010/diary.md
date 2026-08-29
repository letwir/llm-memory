# Diary: ISSUE-008〜010

- timestamp: 2026-08-29 Asia/Tokyo
- task: implement and test remaining issues
- request-evidence: user requested ISSUE-008, ISSUE-009, ISSUE-010 fixes and tests
- action: added optional embedding path, integration concurrency test, and WORM boundary migration/test
- result: concurrency test PASS; WORM test skipped because migration is not applied; embedding live test not run
- friction: pgvector and WORM are intentionally optional DB migrations and require user-side application
- attribution: no prompt defect; remaining uncertainty is environment-bound
- impact: Issues.md remains until live migrations and semantic retrieval are verified
- feedback: apply migrations 008 and 010, then rerun the documented tests
- rewritten-request: verify the optional embedding and WORM migrations on PostgreSQL, rerun integration tests, and close only verified issues
