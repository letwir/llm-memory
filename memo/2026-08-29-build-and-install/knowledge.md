# Knowledge: build and install batch

- task_id: 2026-08-29-build-and-install
- timestamp: 2026-08-29 Asia/Tokyo
- target environment: Windows workspace `A:\Users\letwir\repo\llm-memory`
- evidence/source scope: `BUILD_AND_INSTALL.bat`, `build.ps1`, `SKILL.md`, existing install paths
- verification status: batch guard path and `go test ./...` passed; real distribution was not run
- remaining uncertainty: actual copy permissions and locked executables at each user profile path require user-run verification

`BUILD_AND_INSTALL.bat` builds through `pwsh.exe`, stages the executable, and distributes both `SKILL.md` and `llm-mem.exe` to Gemini, Codex, Claude, `.agents`, and both OpenCode skill paths. It requires `LLM_MEMORY_BUILD_DB_URL` and never prints the value.
