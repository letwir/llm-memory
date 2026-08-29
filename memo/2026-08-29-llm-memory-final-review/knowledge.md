# Knowledge: llm-memory final review

- task_id: 2026-08-29-llm-memory-final-review
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace `A:\Users\letwir\repo\llm-memory`; PostgreSQL 18 remote database
- source scope: official JITMIND/Graphiti licenses, repository source, PostgreSQL migrations,実DB verification output
- verification status: JITMIND MIT and Graphiti Apache-2.0 confirmed; pgvector migration, Gemini embedding save, semantic retrieval, concurrent insert test, and WORM trigger test passed
- completion: all tracked Issues were marked complete, final Issues.md was ingested and deleted, commit `67ef57a` created, worktree clean

JITMIND is MIT and Graphiti is Apache-2.0. This project only uses design inspiration and includes no copied upstream code, so its MIT license does not conflict. If Graphiti code is later copied, Apache notices, license text, and patent terms must be preserved.

The semantic path uses optional pgvector with `gemini-embedding-001` at 768 dimensions. Documents use `RETRIEVAL_DOCUMENT`, queries use `RETRIEVAL_QUERY`; changing model or dimensions requires re-embedding.

The production database has a trigger-maintained `tsvector`, HNSW vector index, normalized active-title uniqueness, and an immutable-field/DELETE guard. Lifecycle updates remain allowed. Real PostgreSQL tests passed for concurrent identity insertion and WORM behavior.
