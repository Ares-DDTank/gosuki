<#
.SYNOPSIS
    重启 (Kill 并 Boot) GoSuki Release 版
.DESCRIPTION
    缺失时自动 build，不缺失时纯启动/纯重启 (Kill 并 Boot)
#>
[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev",
    [string]$Listen = "127.0.0.1:22025",
    [switch]$Rebuild,
    [switch]$StopOnly
)

$ErrorActionPreference = "Stop"
$repoRoot = $PSScriptRoot
$releaseExe = Join-Path $repoRoot "build\gosuki.exe"
$sukiExe = Join-Path $repoRoot "build\suki.exe"
$configGuard = Join-Path $repoRoot "assert-debug-config.ps1"
$resolvedConfigHome = & $configGuard -ConfigHome $ConfigHome

# 1. 终止正在运行的 gosuki 进程
$procs = @(Get-Process -ErrorAction SilentlyContinue | Where-Object {
    try {
        $_.Path -in @($releaseExe, (Join-Path $repoRoot "build\gosuki-debug.exe"))
    } catch { $false }
})
foreach ($p in $procs) {
    Stop-Process -Id $p.Id -Force -ErrorAction SilentlyContinue
}
Start-Sleep -Milliseconds 300

if ($StopOnly) {
    Write-Host "[✓] GoSuki 已停止。" -ForegroundColor Green
    exit 0
}

# 2. 检查 Release 产物，缺失时自动 build
$missing = (-not (Test-Path -LiteralPath $releaseExe)) -or (-not (Test-Path -LiteralPath $sukiExe))
if ($missing -or $Rebuild) {
    Write-Host "[*] GoSuki release 产物缺失或请求重新构建，正在编译..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Path (Join-Path $repoRoot "build") -Force | Out-Null
    Push-Location $repoRoot
    try {
        $version = (& git describe --tags --dirty --always).Trim()
        & go build -tags "windows amd64" -ldflags "-s -w -X github.com/blob42/gosuki/pkg/build.Describe=$version" -o $releaseExe ./cmd/gosuki
        if ($LASTEXITCODE -ne 0) { throw "gosuki release build failed with code $LASTEXITCODE" }
        & go build -tags "windows amd64" -ldflags "-s -w -X github.com/blob42/gosuki/pkg/build.Describe=$version" -o $sukiExe ./cmd/suki
        if ($LASTEXITCODE -ne 0) { throw "suki release build failed with code $LASTEXITCODE" }
        Write-Host "[✓] GoSuki release 编译完成。" -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

# 3. 启动 (Boot)
$configPath = Join-Path $resolvedConfigHome "config.toml"
$databasePath = Join-Path $resolvedConfigHome "gosuki.db"
$logsPath = Join-Path $resolvedConfigHome "logs"
New-Item -ItemType Directory -Path $logsPath -Force | Out-Null

$stdoutLog = Join-Path $logsPath "gosuki.stdout.log"
$stderrLog = Join-Path $logsPath "gosuki.stderr.log"

Write-Host "[*] 正在启动 GoSuki release..." -ForegroundColor Green
$proc = Start-Process `
    -FilePath $releaseExe `
    -ArgumentList @("--config=`"$configPath`"", "--db=`"$databasePath`"", "--listen=$Listen", "start") `
    -WorkingDirectory $repoRoot `
    -WindowStyle Hidden `
    -RedirectStandardOutput $stdoutLog `
    -RedirectStandardError $stderrLog `
    -PassThru

Start-Sleep -Milliseconds 750
$proc.Refresh()
if ($proc.HasExited) {
    throw "GoSuki release exited during startup with code $($proc.ExitCode). See $stderrLog"
}
Write-Host "[✓] GoSuki release 已启动 (PID $($proc.Id)) 于 http://$Listen" -ForegroundColor Green
