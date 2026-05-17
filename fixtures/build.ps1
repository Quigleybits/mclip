# Build all 9 fixture servers (Windows / PowerShell).
#
# Flag rationale:
#   -trimpath           : reproducible builds, strips local path leaks.
#   -ldflags "-s -w"    : strips symbol + DWARF tables (~30% smaller binaries).
#
# Known issue on Windows: see fixtures/README.md "Windows Defender Application
# Control" — small freshly-emitted Go console binaries are occasionally
# classifier-blocked on this user's machine. The blocks are content-based
# and non-deterministic; the recommended workaround is a Defender exclusion
# for the fixtures/ directory. Iterative rebuilds also work but are flaky.
$ErrorActionPreference = 'Stop'
$servers = @(
    'fx-echo','fx-flag-collisions','fx-destructive','fx-resources-watch',
    'fx-http-auth','fx-http-error-data','fx-pagination','fx-prompts','fx-progress'
)
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
foreach ($s in $servers) {
    $dir = Join-Path $root "servers/$s"
    Write-Host "==> $s"
    Push-Location $dir
    try { go build -trimpath -ldflags="-s -w" -o "$s.exe" . } finally { Pop-Location }
}
Write-Host "All 9 fixture servers built."
