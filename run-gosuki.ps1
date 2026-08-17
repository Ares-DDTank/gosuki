[CmdletBinding()]
param(
    [ValidateSet("fork", "scoop")]
    [string]$Runtime = "fork",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$CommandArguments
)

$ErrorActionPreference = "Stop"
$runtimeProfile = & (Join-Path $PSScriptRoot "resolve-gosuki-runtime.ps1") -Runtime $Runtime
$reserved = @("--config", "-c", "--db", "--listen", "-l")
foreach ($argument in $CommandArguments) {
    if ($reserved | Where-Object { $argument -eq $_ -or $argument.StartsWith("$_=") }) {
        throw "The runtime wrapper owns '$argument'; select the target with -Runtime instead."
    }
}

$arguments = @(
    "--config=$($runtimeProfile.ConfigPath)"
    "--db=$($runtimeProfile.DatabasePath)"
    "--listen=$($runtimeProfile.Listen)"
) + $CommandArguments

& $runtimeProfile.GosukiExecutable @arguments
exit $LASTEXITCODE
