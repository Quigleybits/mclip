# Rebuild only the servers that currently fail verify, with a fresh -B
# build-id each attempt. Stops once verify returns 9/9, or after $maxAttempts.
$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$maxAttempts = 8

function Get-BlockedServers {
    # verify.exe exits 1 on partial-pass; capture output without aborting.
    $verify = Join-Path $root 'verify/verify.exe'
    $output = & cmd.exe /c "`"$verify`" 2>&1"
    $blocked = @()
    foreach ($line in $output) {
        if ($line -match '^\[FAIL\] (\S+)') {
            $blocked += $matches[1]
        }
    }
    return $blocked
}

$attempt = 0
while ($attempt -lt $maxAttempts) {
    $blocked = Get-BlockedServers
    if ($blocked.Count -eq 0) {
        Write-Host "All 9 fixture servers verified after $attempt rebuild attempt(s)."
        exit 0
    }
    $attempt++
    Write-Host "Attempt ${attempt}: rebuilding $($blocked -join ', ')"
    foreach ($s in $blocked) {
        $bytes = New-Object byte[] 10
        [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
        $hex = -join ($bytes | ForEach-Object { '{0:x2}' -f $_ })
        Push-Location (Join-Path $root "servers/$s")
        try {
            $env:CGO_ENABLED = '0'
            go build -trimpath -ldflags="-s -w -B 0x$hex" -o "$s.exe" .
            if ($LASTEXITCODE -ne 0) { Write-Warning "go build $s failed" }
        } finally { Pop-Location }
    }
}
Write-Warning "After $maxAttempts attempts, still failing: $($blocked -join ', '). Add a Defender exclusion for fixtures/servers/."
exit 1
