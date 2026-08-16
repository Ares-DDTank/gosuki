[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev",
    [string]$Listen = "127.0.0.1:22025",
    [string]$DebugLevel = "info,chrome=debug",
    [switch]$SkipTests
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$configGuard = Join-Path $repoRoot "assert-debug-config.ps1"
$buildScript = Join-Path $repoRoot "build-debug.ps1"
$startScript = Join-Path $repoRoot "start-debug.ps1"

& $configGuard -ConfigHome $ConfigHome | Out-Null
& $buildScript -SkipTests:$SkipTests | Out-Null
& $startScript -ConfigHome $ConfigHome -Listen $Listen -DebugLevel $DebugLevel
