# Проверка техкарт: версия, ответственный, тестовый пример.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

$fail = $false
Get-ChildItem "src/rules/techcards" -Filter *.json | ForEach-Object {
    $tc = Get-Content -Raw $_.FullName | ConvertFrom-Json
    if (-not $tc.version) { Write-Host "FAIL: $($_.Name) без version"; $fail = $true }
    if (-not $tc.owner) { Write-Host "FAIL: $($_.Name) без owner"; $fail = $true }
    if (-not $tc.test_examples -or $tc.test_examples.Count -eq 0) {
        Write-Host "FAIL: $($_.Name) без test_examples"; $fail = $true
    } else {
        foreach ($ex in $tc.test_examples) {
            if (-not (Test-Path -LiteralPath $ex)) {
                Write-Host "FAIL: $($_.Name): файл test_examples не найден: $ex"; $fail = $true
            }
        }
    }
}
if ($fail) { exit 1 } else { Write-Host "techcards ok"; exit 0 }
