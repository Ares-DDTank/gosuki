[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev",
    [string]$Listen = "127.0.0.1:22025",
    [string]$DebugLevel = "info,chrome=debug"
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$debugExe = [IO.Path]::GetFullPath((Join-Path $repoRoot "build\gosuki-debug.exe"))
$configGuard = Join-Path $repoRoot "assert-debug-config.ps1"
$resolvedConfigHome = & $configGuard -ConfigHome $ConfigHome

if ($Listen -eq "127.0.0.1:2025") {
    throw "Port 2025 is reserved for Scoop GoSuki; use the isolated debug port instead."
}

if (-not (Test-Path -LiteralPath $debugExe -PathType Leaf)) {
    throw "Missing debug executable: $debugExe. Run .\build-debug.ps1 first."
}

$running = @(
    Get-Process -ErrorAction SilentlyContinue |
        Where-Object {
            try {
                [string]::Equals($_.Path, $debugExe, [StringComparison]::OrdinalIgnoreCase)
            }
            catch {
                $false
            }
        }
)
if ($running.Count -gt 0) {
    throw "GoSuki debug is already running with PID(s): $($running.Id -join ', ')."
}

$configPath = Join-Path $resolvedConfigHome "config.toml"
$databasePath = Join-Path $resolvedConfigHome "gosuki.db"
$importsPath = Join-Path $resolvedConfigHome "imports"
$logsPath = Join-Path $resolvedConfigHome "logs"
New-Item -ItemType Directory -Path $importsPath, $logsPath -Force | Out-Null

$stdoutLog = Join-Path $logsPath "gosuki-debug.stdout.log"
$stderrLog = Join-Path $logsPath "gosuki-debug.stderr.log"
$arguments = @(
    "--config=`"$configPath`""
    "--db=`"$databasePath`""
    "--listen=$Listen"
    "--debug=$DebugLevel"
    "start"
)

$process = Start-Process `
    -FilePath $debugExe `
    -ArgumentList $arguments `
    -WorkingDirectory $repoRoot `
    -WindowStyle Hidden `
    -RedirectStandardOutput $stdoutLog `
    -RedirectStandardError $stderrLog `
    -PassThru

Start-Sleep -Milliseconds 750
$process.Refresh()
if ($process.HasExited) {
    throw "GoSuki debug exited during startup with code $($process.ExitCode). See $stderrLog"
}

Write-Host "Started GoSuki debug PID $($process.Id) at http://$Listen"
$process
