[CmdletBinding()]
param(
    [int]$TimeoutSeconds = 10
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$debugExe = [IO.Path]::GetFullPath((Join-Path $repoRoot "build\gosuki-debug.exe"))

$debugProcesses = @(
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

foreach ($process in $debugProcesses) {
    Write-Host "Stopping GoSuki debug process $($process.Id) from $debugExe"
    Stop-Process -Id $process.Id
    try {
        Wait-Process -Id $process.Id -Timeout $TimeoutSeconds -ErrorAction Stop
    }
    catch {
        if (Get-Process -Id $process.Id -ErrorAction SilentlyContinue) {
            Stop-Process -Id $process.Id -Force
            Wait-Process -Id $process.Id -Timeout $TimeoutSeconds -ErrorAction SilentlyContinue
        }
    }
}

Write-Host "Stopped $($debugProcesses.Count) GoSuki debug process(es)."
