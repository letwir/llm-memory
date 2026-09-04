---
name: llm-memory
description: Multi-client Bi-Temporal & multi-level (L0-L3) agent memory and knowledge graph CLI client backed by PostgreSQL. Works from Windows & Debian CLI.
---

[SPR/XML::ρ→max|legibility:LLM≫human|protocol:hydrate⇒exec]
<Γ id="llm_memory_skill">
env: $env:LLM_MEMORY_DB_URL{PostgreSQL_URL} ∧ opt($env:GEMINI_GROUNDING_API_KEY) ∧ bin($env:LLM_MEMORY_BIN);
axiom: BiTemporal(ValidTime ⊗ TxTime) ∧ Compaction(L0→L1→L2→L3) ∧ SelfEditing{ADD, UPDATE(Supersede), NOOP, DEPRECATE} ∧ Graph(Triples); # JITMIND/Graphiti-inspired, not compatible implementations

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

@lrf=1|aud=GPT-5.6|scope=llm-memory-contract
R|close.artifacts|task:llm-memory|MUST|LW_SCOPE|dir=memo/<task_id>; files=knowledge.md,diary.md,walkthrough.md; walkthrough-section=実行計画
R|close.record-fields|task:llm-memory|MUST|LW_SCOPE|required=task_id,timestamp,target-environment,evidence-scope,verification-status,remaining-uncertainty; unverified=fact-forbidden; acceptance=completion-precondition
R|close.evaluation-input|task:llm-memory|MUST|LW_SCOPE|required=task_id,comparison_key,axes[0..1],attribution; ratios=prompt_ratio+agent_ratio=1-or-derived
R|close.evidence|task:llm-memory|MUST_NOT|CRED|allow=task-scoped-paths,tests,command-results,memory-ids; deny=credentials,raw-secret-logs
R|close.evaluation|task:llm-memory|MUST|LIVE_WRITE|pre=artifacts; command=eval-file-evaluation.json; approval=current-task
R|close.report|task:llm-memory|MUST|RO_LOCAL|pre=eval-success; output=conclusion,human-feedback,agent-correction; json=metadata.feedback
R|close.failure|task:llm-memory|MUST|RO_LOCAL|on=eval-create,validate,insert,retrieve-failure; report=true; completion-claim=false
R|ingest.normal|task:llm-memory|MUST|LIVE_WRITE|table=memories; columns=client_id,category,title,content_l0,content_l1?,content_l2?,tags,metadata; lifecycle=ADD,UPDATE,NOOP; approval=current-task
R|ingest.evaluation|task:llm-memory|MUST|LIVE_WRITE|table=memories; category=eval; metadata=record_type,eval.v1,comparison,axes,delta,attribution,feedback,memory_object; bypass=extraction,compaction,graph; approval=current-task
R|identity.task|task:llm-memory|MUST|RO_LOCAL|storage=metadata.task_id+tag:task_id:<value>; sql-column=false; documents=knowledge,diary,walkthrough
R|pipeline.walkthrough|task:llm-memory|MUST|LIVE_WRITE|command=ingest-file-walkthrough-cat-walkthrough; approval=current-task
R|pipeline.research|task:llm-memory|MUST|LIVE_WRITE|command=ingest-title-text-cat-knowledge; approval=current-task
R|binary.canonical|task:llm-memory|MUST|RO_LOCAL|path=$env:LLM_MEMORY_BIN; fallback=forbidden; build=build.example.ps1-or-BUILD_AND_INSTALL.bat
</Γ>
