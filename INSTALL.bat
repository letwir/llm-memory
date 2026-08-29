@echo off
setlocal EnableExtensions

set "ROOT=%~dp0"
set "SRC=%ROOT%SKILL.md"

if not exist "%SRC%" (
    echo ERROR: SKILL.md was not found next to INSTALL.bat.
    exit /b 1
)

rem Make the project location discoverable by installed skills.
setx LLM_MEMORY_HOME "%ROOT%" >nul
setx LLM_MEMORY_BIN "%ROOT%llm-mem.exe" >nul

set "TARGET=%~1"
if "%TARGET%"=="" set "TARGET=all"

if /I "%TARGET%"=="all" (
    call :install "%USERPROFILE%\.gemini\skills\llm-memory"
    call :install "%USERPROFILE%\.codex\skills\llm-memory"
    call :install "%USERPROFILE%\.claude\skills\llm-memory"
    call :install "%USERPROFILE%\.agents\skills\llm-memory"
    call :install "%USERPROFILE%\.config\opencode\skills\llm-memory"
    call :install "%USERPROFILE%\.opencode\skills\llm-memory"
    goto :done
)

if /I "%TARGET%"=="gemini" (
    call :install "%USERPROFILE%\.gemini\skills\llm-memory"
    goto :done
)
if /I "%TARGET%"=="codex" (
    call :install "%USERPROFILE%\.codex\skills\llm-memory"
    goto :done
)
if /I "%TARGET%"=="claude" (
    call :install "%USERPROFILE%\.claude\skills\llm-memory"
    goto :done
)
if /I "%TARGET%"=="agents" (
    call :install "%USERPROFILE%\.agents\skills\llm-memory"
    goto :done
)
if /I "%TARGET%"=="opencode" (
    call :install "%USERPROFILE%\.config\opencode\skills\llm-memory"
    call :install "%USERPROFILE%\.opencode\skills\llm-memory"
    goto :done
)

echo Usage: INSTALL.bat [all^|gemini^|codex^|claude^|agents^|opencode]
exit /b 2

:install
set "DEST=%~1"
if not exist "%DEST%" mkdir "%DEST%"
if not exist "%DEST%" (
    echo ERROR: Could not create %DEST%
    exit /b 1
)
if exist "%DEST%\SKILL.md" if not exist "%DEST%\SKILL.md.previous" copy /Y "%DEST%\SKILL.md" "%DEST%\SKILL.md.previous" >nul
copy /Y "%SRC%" "%DEST%\SKILL.md" >nul
if errorlevel 1 (
    echo ERROR: Failed to install skill into %DEST%
    exit /b 1
)
echo Installed llm-memory skill into %DEST%
exit /b 0

:done
echo.
echo Installation complete. Open a new terminal so LLM_MEMORY_BIN is refreshed.
echo Build the binary before using the skill: powershell -ExecutionPolicy Bypass -File build.example.ps1
exit /b 0
