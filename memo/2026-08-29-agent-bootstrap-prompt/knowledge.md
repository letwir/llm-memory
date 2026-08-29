# Knowledge: agent bootstrap prompt

- task_id: 2026-08-29-agent-bootstrap-prompt
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace `A:\Users\letwir\repo\llm-memory`
- evidence/source scope: README.md, USAGE.md, llm-memory SKILL.md, CLI test
- verification status: prompt documented; `go test ./...` passed with repository-local GOCACHE
- remaining uncertainty: prompt behavior depends on each agent honoring injected instructions

README now includes a one-line SPR/XML-style bootstrap instruction that invokes llm-memory and requires Explore, Plan, Implement, Verify, and Report phases. It explicitly requests initial scope/constraint/file/state exploration and final changes/tests/uncertainty/next-action reporting.
