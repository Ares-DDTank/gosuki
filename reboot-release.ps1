[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev",
    [string]$Listen = "127.0.0.1:22025",
    [switch]$Rebuild,
    [switch]$StopOnly
)
& "$PSScriptRoot\restart-release.ps1" @PSBoundParameters
