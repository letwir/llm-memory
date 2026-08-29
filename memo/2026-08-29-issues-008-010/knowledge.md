# Knowledge: ISSUE-008〜010

- task_id: 2026-08-29-issues-008-010
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace and configured PostgreSQL 18 remote database
- evidence/source scope: migrations/008_embeddings.sql, migrations/010_worm_boundary.sql, embedding.go, integration_test.go, user database schema
- verification status: local tests and concurrent DB test passed; WORM trigger and pgvector/embedding live paths remain unverified
- remaining uncertainty: pgvector availability, Gemini embedding API quota/permissions, and live trigger application
