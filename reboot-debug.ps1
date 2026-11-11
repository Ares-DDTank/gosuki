<#
.SYNOPSIS
    重启 (Kill 并 Boot) GoSuki Debug 版
.DESCRIPTION
    无条件强制重新编译 Debug 版后再重启 (Kill 并 Boot)
#>
[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev",
    [string]$Listen = "127.0.0.1:22025",
    [string]$DebugLevel = "info,chrome=debug",
    [switch]$SkipTests
)

& "$PSScriptRoot\restart-debug.ps1" @PSBoundParameters