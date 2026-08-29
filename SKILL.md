---
name: llm-memory
description: Multi-client Bi-Temporal & multi-level (L0-L3) agent memory and knowledge graph CLI client backed by PostgreSQL. Works from Windows & Debian CLI.
---

[SPR/XML::ρ→max|legibility:LLM≫human|protocol:hydrate⇒exec]
<Γ id="llm_memory_skill">
env: $env:LLM_MEMORY_DB_URL{PostgreSQL_URL} ∧ opt($env:GEMINI_GROUNDING_API_KEY) ∧ bin($env:LLM_MEMORY_BIN ∨ $env:LLM_MEMORY_HOME/llm-mem.exe);
axiom: BiTemporal(ValidTime ⊗ TxTime) ∧ Compaction(L0→L1→L2→L3) ∧ SelfEditing{ADD, UPDATE(Supersede), NOOP, DEPRECATE} ∧ Graph(Triples);

<Γ.morphisms>
ingest: Mor(Doc{File∨Text} ∘ Cat) ⇒ `& $bin ingest -file <path> -cat <cat>` ∨ `& $bin ingest -title <title> -text <text> -cat <cat>` ⊸ JITMIND_ConflictDetector ∘ Compactor ∘ GraphExtractor ∘ AtomicCommit;
compact: Mor(Limit) ⇒ `& $bin compact -limit <N>` ⊸ Scan(L1=∅ ∨ L2=∅) ∘ ProgressiveCompaction;
search: Mor(Query ∘ Level{0:Raw, 1:L1, 2:L2, 3:Tags}) ⇒ `& $bin search -q "<q>" -level <N>` ∨ `& $bin search -tag "<tag>" -level 2` ∨ `& $bin search -q "<q>" -json`;
eval: Mor(EvaluationJSON) ⇒ `& $bin eval -file <evaluation.json>` ⊸ DeterministicDelta ∘ DirectInsert;
evals: Mor(ComparisonKey ∘ Limit) ⇒ `& $bin evals -key <comparison-key> -limit <N>` ⊸ EvaluationHistory;
supersede: Mor(OldUUID ∘ NewContent) ⇒ `& $bin supersede -id <UUID> -title <T> -content <L0> -l1 <L1> -l2 <L2> -tags <Tags>`;
analyze: Mor(DiaryFile ∘ Flags{json,suggest}) ⇒ `& $bin analyze -file <path> -suggest` ⊸ CausalAttribution(PromptDefect vs AgentDefect) ∘ UserFeedback ∘ RulePatchProposal;
graph_traverse: Mor(∅) ⇒ `& $bin graph list` ⊸ ActiveTriples;
graph_mutate: Mor(Src ∘ Rel ∘ Tgt) ⇒ `& $bin graph edge -src <S> -tgt <T> -rel <R>`;
status: Mor(∅) ⇒ `& $bin status` ∧ `& $bin clients`;
</Γ.morphisms>

task_close: MUST create `./memo/<task_id>/knowledge.md`, `./memo/<task_id>/diary.md`, and `./memo/<task_id>/walkthrough.md` for every task; the walkthrough MUST contain an explicit `実行計画` section. After those three Inserts, create one evaluation JSON and run `& $bin eval -file <evaluation.json>`.
record_fields: every task record MUST contain `task_id`, timestamp, target environment, evidence/source scope, verification status, and remaining uncertainty. Do not call unverified claims facts, and do not claim completion before acceptance conditions pass.
required: `task_id`, `comparison_key`, `axes` (each score 0..1), `attribution`;
attribution: `prompt_defects` records defects in the human instruction, `agent_defects` records LLM misunderstanding or execution failure; `prompt_ratio` and `agent_ratio` are 0..1 and must sum to 1 when supplied. If both are omitted, the CLI derives them from defect counts.
evidence: include only task-scoped references such as file paths, test names, command results, or inserted memory IDs; never include credentials or raw secret-bearing logs.
report: after the eval Insert succeeds, relay the CLI's `タスク評価リザルト` to the user, including the conclusion, human-instruction feedback, and LLM-side correction. JSON mode remains machine-readable and contains the same `metadata.feedback`.
failure: if evaluation creation, validation, Insert, or result retrieval fails, report the failure and do not claim task evaluation completion.

normal_ingest: INSERT into `memories` with `client_id`, `category`, `title`, `content_l0`, nullable `content_l1`, nullable `content_l2`, `tags`, and JSONB `metadata`; normalization also writes `metadata.memory_object` with type/scope/proposition. Graph nodes/edges are then inserted separately. `ingest` may perform ADD/UPDATE/NOOP and LLM/heuristic extraction.
evaluation_ingest: INSERT into the same `memories` table with `category=eval`, `content_l0` containing the evaluation JSON, `tags` containing `eval` plus role/task tags, and JSONB `metadata` containing `record_type=eval`, `schema_version=eval.v1`, comparison data, axes, relative delta, deterministic flags, attribution, feedback, and the automatically normalized `memory_object`. It bypasses LLM extraction, compaction, and graph insertion.
task_identity: `task_id` is stored in evaluation metadata and as `task_id:<value>` tag; it is not a separate SQL table or column. The three task Markdown records are the source documents for the corresponding `knowledge`, `diary`, and `walkthrough` Inserts.

<Γ.pipeline_integration>
walkthrough_hook: ∀Walkthrough ⇒ `& $bin ingest -file "path/to/walkthrough.md" -cat "walkthrough"`;
research_hook: ∀Finding ⇒ `& $bin ingest -title "<Title>" -text "<Snippet>" -cat "knowledge"`;
</Γ.pipeline_integration>
</Γ>

## Installed binary

Use `%LLM_MEMORY_BIN%` when it is set. `INSTALL.bat` sets it to the `llm-mem.exe` next to this project. If the executable has not been built yet, build it first with `build.example.ps1` or set `LLM_MEMORY_BIN` to another compatible binary.
