[CmdletBinding()]
param(
    [string]$ConfigHome = "D:\C2D\dotfiles\gosuki-dev"
)

$ErrorActionPreference = "Stop"
$expectedRole = "gosuki-fork"
$markerPath = Join-Path $ConfigHome ".gosuki-config-role"
$configPath = Join-Path $ConfigHome "config.toml"

if (-not (Test-Path -LiteralPath $ConfigHome -PathType Container)) {
    throw "Missing GoSuki fork config directory: $ConfigHome"
}

if (-not (Test-Path -LiteralPath $markerPath -PathType Leaf)) {
    throw "Missing GoSuki config role marker: $markerPath"
}

$actualRole = (Get-Content -LiteralPath $markerPath -Raw).Trim()
if ($actualRole -ne $expectedRole) {
    throw "Refusing to start gosuki-fork with config role '$actualRole' from $ConfigHome; expected '$expectedRole'."
}

if (-not (Test-Path -LiteralPath $configPath -PathType Leaf)) {
    throw "Missing GoSuki fork config file: $configPath"
}

$resolvedHome = (Resolve-Path -LiteralPath $ConfigHome).Path
$productionConfig = Join-Path $env:APPDATA "gosuki\config.toml"
if (Test-Path -LiteralPath $productionConfig -PathType Leaf) {
    $resolvedDevConfig = (Resolve-Path -LiteralPath $configPath).Path
    $resolvedProductionConfig = (Resolve-Path -LiteralPath $productionConfig).Path
    if ([string]::Equals($resolvedDevConfig, $resolvedProductionConfig, [StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to use the Scoop GoSuki production config for a debug launch: $resolvedDevConfig"
    }
}

$resolvedHome
