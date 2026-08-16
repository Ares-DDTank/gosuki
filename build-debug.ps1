[CmdletBinding()]
param(
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$debugExe = Join-Path $repoRoot "build\gosuki-debug.exe"
$stopScript = Join-Path $repoRoot "stop-debug.ps1"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not available on PATH."
}

& $stopScript
New-Item -ItemType Directory -Path (Split-Path -Parent $debugExe) -Force | Out-Null

Push-Location $repoRoot
try {
    if (-not $SkipTests) {
        & go test -count=1 ./browsers/chrome
        if ($LASTEXITCODE -ne 0) {
            throw "Chrome tests failed with exit code $LASTEXITCODE."
        }
    }

    $version = (& git describe --tags --dirty --always).Trim() + "-debug"
    & go build `
        -tags "windows amd64" `
        -gcflags "all=-N -l" `
        -ldflags "-X github.com/blob42/gosuki/pkg/build.Describe=$version" `
        -o $debugExe `
        ./cmd/gosuki
    if ($LASTEXITCODE -ne 0) {
        throw "GoSuki debug build failed with exit code $LASTEXITCODE."
    }
}
finally {
    Pop-Location
}

$item = Get-Item -LiteralPath $debugExe
Write-Host "Built GoSuki debug executable: $($item.FullName) ($($item.Length) bytes)"
$item.FullName
