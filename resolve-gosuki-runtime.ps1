[CmdletBinding()]
param(
    [ValidateSet("fork", "scoop")]
    [string]$Runtime = "fork"
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot

if ($Runtime -eq "fork") {
    $configHome = "D:\C2D\dotfiles\gosuki-dev"
    $expectedRole = "gosuki-fork"
    $binaryRoot = Join-Path $repoRoot "build"
    $gosukiExe = Join-Path $binaryRoot "gosuki-debug.exe"
    $sukiExe = Join-Path $binaryRoot "suki-debug.exe"
    $databasePath = Join-Path $configHome "gosuki.db"
    $listen = "127.0.0.1:22025"
}
else {
    $configHome = "D:\C2D\dotfiles\gosuki"
    $expectedRole = "scoop-original"
    $binaryRoot = (& scoop prefix gosuki).Trim()
    if (-not $binaryRoot) {
        throw "Could not resolve the Scoop GoSuki installation."
    }
    $gosukiExe = Join-Path $binaryRoot "gosuki.exe"
    $sukiExe = Join-Path $binaryRoot "suki.exe"
    $databasePath = Join-Path $env:APPDATA "gosuki\gosuki.db"
    $listen = "127.0.0.1:2025"
}

$markerPath = Join-Path $configHome ".gosuki-config-role"
$configPath = Join-Path $configHome "config.toml"
foreach ($requiredPath in @($markerPath, $configPath, $gosukiExe, $sukiExe)) {
    if (-not (Test-Path -LiteralPath $requiredPath -PathType Leaf)) {
        throw "Missing $Runtime runtime file: $requiredPath"
    }
}

$actualRole = (Get-Content -LiteralPath $markerPath -Raw).Trim()
if ($actualRole -ne $expectedRole) {
    throw "Runtime '$Runtime' has config role '$actualRole'; expected '$expectedRole'."
}

[pscustomobject]@{
    Runtime = $Runtime
    ConfigHome = (Resolve-Path -LiteralPath $configHome).Path
    ConfigPath = (Resolve-Path -LiteralPath $configPath).Path
    DatabasePath = [IO.Path]::GetFullPath($databasePath)
    Listen = $listen
    GosukiExecutable = [IO.Path]::GetFullPath($gosukiExe)
    SukiExecutable = [IO.Path]::GetFullPath($sukiExe)
}
