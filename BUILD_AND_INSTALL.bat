@echo off
setlocal EnableExtensions

set "ROOT=%~dp0"
set "TEMP_BIN=%ROOT%.llm-mem.deploy.%RANDOM%.exe"

if "%LLM_MEMORY_BUILD_DB_URL%"=="" (
    echo ERROR: LLM_MEMORY_BUILD_DB_URL is not set.
    echo Set it in the current terminal, then run this batch again.
    exit /b 2
)

where pwsh.exe >nul 2>&1
if errorlevel 1 (
    echo ERROR: pwsh.exe was not found in PATH.
    exit /b 3
)
where go.exe >nul 2>&1
if errorlevel 1 (
    echo ERROR: go.exe was not found in PATH.
    exit /b 3
)

echo Building llm-mem.exe...
pwsh.exe -NoProfile -ExecutionPolicy Bypass -File "%ROOT%build.ps1" -Output "%TEMP_BIN%"
if errorlevel 1 goto :fail
if not exist "%TEMP_BIN%" (
    echo ERROR: Build reported success but the binary was not created.
    goto :fail
)

copy /Y "%TEMP_BIN%" "%ROOT%llm-mem.exe" >nul
if errorlevel 1 goto :fail

echo Distributing SKILL.md and llm-mem.exe to each skill directory...
call :deploy "%USERPROFILE%\.gemini\skills\llm-memory"
if errorlevel 1 goto :fail
call :deploy "%USERPROFILE%\.codex\skills\llm-memory"
if errorlevel 1 goto :fail
call :deploy "%USERPROFILE%\.claude\skills\llm-memory"
if errorlevel 1 goto :fail
call :deploy "%USERPROFILE%\.agents\skills\llm-memory"
if errorlevel 1 goto :fail
call :deploy "%USERPROFILE%\.config\opencode\skills\llm-memory"
if errorlevel 1 goto :fail
call :deploy "%USERPROFILE%\.opencode\skills\llm-memory"
if errorlevel 1 goto :fail

setx LLM_MEMORY_HOME "%ROOT%" >nul
setx LLM_MEMORY_BIN "%ROOT%llm-mem.exe" >nul
del /Q "%TEMP_BIN%" >nul 2>&1
echo.
echo Build and distribution complete.
echo Open a new terminal so LLM_MEMORY_BIN is refreshed.
exit /b 0

:deploy
if not exist "%~1" mkdir "%~1"
if not exist "%~1" (
    echo ERROR: Could not create %~1
    exit /b 1
)
copy /Y "%ROOT%SKILL.md" "%~1\SKILL.md" >nul
if errorlevel 1 (
    echo ERROR: Failed to deploy SKILL.md to %~1
    exit /b 1
)
copy /Y "%ROOT%llm-mem.exe" "%~1\llm-mem.exe" >nul
if errorlevel 1 (
    echo ERROR: Failed to deploy binary to %~1
    exit /b 1
)
echo Deployed binary to %~1
exit /b 0

:fail
echo ERROR: Build or distribution failed. Existing deployed binaries were not intentionally removed.
del /Q "%TEMP_BIN%" >nul 2>&1
exit /b 1
