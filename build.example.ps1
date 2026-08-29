param(
    [string]$Output = "llm-mem.exe"
)

$ErrorActionPreference = "Stop"

# この値はexeへ埋め込まれます。公開リポジトリへ値を書き込まないでください。
$DatabaseURL = $env:LLM_MEMORY_BUILD_DB_URL
if ([string]::IsNullOrWhiteSpace($DatabaseURL)) {
    throw "LLM_MEMORY_BUILD_DB_URL を設定してください。"
}

$ldflags = "-X main.buildDatabaseURL=$DatabaseURL"
go build -trimpath -ldflags $ldflags -o $Output .
Write-Output "built $Output with an embedded database destination"
