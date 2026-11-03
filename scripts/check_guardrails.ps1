# Проверка, что guardrails не ослаблены.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

$fail = $false

$guardrails = Get-Content -Raw "src/harness/guardrails/__init__.py"
if ($guardrails -notmatch "def check_acl") { Write-Host "FAIL: check_acl отсутствует"; $fail = $true }
if ($guardrails -notmatch "def check_write_allowed") { Write-Host "FAIL: check_write_allowed отсутствует"; $fail = $true }

$agents = Get-Content -Raw "AGENTS.md"
if ($agents -notmatch "code-based evals") { Write-Host "FAIL: AGENTS.md не фиксирует code-based ACL"; $fail = $true }

if ($fail) { exit 1 } else { Write-Host "guardrails ok"; exit 0 }
