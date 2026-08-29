# Knowledge: docs and Mermaid synchronization

- task_id: 2026-08-29-docs-mermaid-sync
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace `A:\Users\letwir\repo\llm-memory`
- evidence/source scope: README/USAGE/SKILL, docs Mermaid diagrams, init.sql and migrations
- verification status: docs synchronized with pgvector, search_document, WORM trigger, NULL valid_to, and BUILD_AND_INSTALL; Go tests passed
- remaining uncertainty: GitHub rendering itself was not browser-captured; syntax was simplified to GitHub-supported fenced Mermaid blocks

The documentation now reflects the implemented optional semantic path, trigger-maintained full-text search, active-title identity, WORM boundary, and self-contained build/distribution batch. Mermaid diagrams avoid compound node arrows and stale database semantics.
