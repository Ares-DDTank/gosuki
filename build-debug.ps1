[CmdletBinding()]
param(
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$debugExecutables = @(
    [pscustomobject]@{ Name = "GoSuki"; Path = (Join-Path $repoRoot "build\gosuki-debug.exe"); Package = "./cmd/gosuki" }
    [pscustomobject]@{ Name = "Suki"; Path = (Join-Path $repoRoot "build\suki-debug.exe"); Package = "./cmd/suki" }
)
$stopScript = Join-Path $repoRoot "stop-debug.ps1"

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    throw "Go is not available on PATH."
}

& $stopScript
New-Item -ItemType Directory -Path (Join-Path $repoRoot "build") -Force | Out-Null

Push-Location $repoRoot
try {
    if (-not $SkipTests) {
        & go test -count=1 ./browsers/chrome
        if ($LASTEXITCODE -ne 0) {
            throw "Chrome tests failed with exit code $LASTEXITCODE."
        }
    }

    $version = (& git describe --tags --dirty --always).Trim() + "-debug"
    foreach ($target in $debugExecutables) {
        & go build `
            -tags "windows amd64" `
            -gcflags "all=-N -l" `
            -ldflags "-X github.com/blob42/gosuki/pkg/build.Describe=$version" `
            -o $target.Path `
            $target.Package
        if ($LASTEXITCODE -ne 0) {
            throw "$($target.Name) debug build failed with exit code $LASTEXITCODE."
        }
    }
}
finally {
    Pop-Location
}

foreach ($target in $debugExecutables) {
    $item = Get-Item -LiteralPath $target.Path
    Write-Host "Built $($target.Name) debug executable: $($item.FullName) ($($item.Length) bytes)"
    $item.FullName
}
