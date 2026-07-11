#Requires -Version 7
<#
.SYNOPSIS
    The ONE canonical verification gate. Humans, the Makefile shim, and CI all
    run this same script - do not restate these commands anywhere else.

.DESCRIPTION
    Ordered stages: restore (locked) -> format-check -> build (warnings-as-
    errors, Release) -> test (unit) -> audit (vulnerable packages, direct +
    transitive). The script stops at the first failing stage and exits
    non-zero. PowerShell 7 (pwsh) is the only blessed script runtime: one
    script, identical behavior on Windows, Linux, and macOS.

    Policy docs: csharp/quality/linting.md, csharp/operations/ci-and-release.md

.PARAMETER Stage
    Run a single stage (plus its prerequisites) instead of the full gate.
    Used by the Makefile shim: e.g. `pwsh ./verify.ps1 -Stage build`.

.PARAMETER Integration
    Also run tests/*IntegrationTests* projects. Off by default because they
    need Docker (Testcontainers), which is not guaranteed on every dev machine.

.PARAMETER Fix
    With -Stage format: rewrite files (dotnet format) instead of verifying.

.EXAMPLE
    pwsh ./verify.ps1                     # full gate (what CI runs)
    pwsh ./verify.ps1 -Integration        # full gate + integration tests
    pwsh ./verify.ps1 -Stage format -Fix  # fix formatting in place
#>
[CmdletBinding()]
param(
    [ValidateSet('all', 'restore', 'format', 'build', 'test', 'audit')]
    [string]$Stage = 'all',

    [switch]$Integration,

    [switch]$Fix
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'
# Native exit codes are checked explicitly via $LASTEXITCODE after every
# command; do not let pwsh convert them into terminating errors implicitly.
$PSNativeCommandUseErrorActionPreference = $false

# Always operate from the repo root (the directory containing this script),
# no matter where the caller's shell happens to be.
Set-Location -Path $PSScriptRoot

function Write-Stage {
    param([string]$Name)
    Write-Host ''
    Write-Host "==> $Name" -ForegroundColor Cyan
}

# Every native command is followed by this check; a non-zero exit code stops
# the gate immediately with a message naming the failed stage.
function Assert-LastExitCode {
    param([string]$StageName)
    if ($LASTEXITCODE -ne 0) {
        Write-Host "FAIL: $StageName (exit code $LASTEXITCODE)" -ForegroundColor Red
        exit 1
    }
}

function Invoke-RestoreStage {
    Write-Stage 'restore (locked mode)'
    # --locked-mode fails if packages.lock.json disagrees with the project
    # graph, instead of silently rewriting it. Drift is a finding, not a fix.
    dotnet restore --locked-mode
    Assert-LastExitCode 'restore'
}

function Invoke-FormatStage {
    if ($Fix) {
        Write-Stage 'format (rewrite in place)'
        dotnet format --no-restore
        Assert-LastExitCode 'format'
        return
    }
    Write-Stage 'format-check'
    # Verify-only: CI and the gate never mutate the working tree.
    dotnet format --verify-no-changes --no-restore
    Assert-LastExitCode 'format-check (run: pwsh ./verify.ps1 -Stage format -Fix)'
}

function Invoke-BuildStage {
    Write-Stage 'build (Release, warnings as errors)'
    # -warnaserror backstops TreatWarningsAsErrors even if a project opts out.
    dotnet build -c Release -warnaserror --no-restore
    Assert-LastExitCode 'build'
}

# Unit tests always; tests/*IntegrationTests* only behind -Integration.
# Convention: test projects live under tests/ and integration projects carry
# the .IntegrationTests suffix (csharp/quality/testing.md).
function Get-TestProjects {
    param([bool]$IncludeIntegration)
    $all = @(Get-ChildItem -Path 'tests' -Recurse -Filter '*.csproj' -ErrorAction SilentlyContinue)
    if ($IncludeIntegration) {
        return $all
    }
    return @($all | Where-Object { $_.BaseName -notlike '*IntegrationTests*' })
}

function Invoke-TestStage {
    Write-Stage "test (unit$(if ($Integration) { ' + integration' }))"
    # @() at the CALL SITE: PowerShell unrolls a function's return value, so a
    # single matching project would come back as one FileInfo whose missing
    # .Count property is a hard error under Set-StrictMode.
    $projects = @(Get-TestProjects -IncludeIntegration:$Integration.IsPresent)
    if ($projects.Count -eq 0) {
        Write-Host 'FAIL: no test projects found under tests/' -ForegroundColor Red
        exit 1
    }
    foreach ($project in $projects) {
        Write-Host "--- $($project.BaseName)"
        # --no-build: the build stage already compiled Release. Test projects
        # run on Microsoft.Testing.Platform via xunit.v3.
        dotnet test $project.FullName -c Release --no-build
        Assert-LastExitCode "test ($($project.BaseName))"
    }
}

function Invoke-AuditStage {
    Write-Stage 'audit (vulnerable packages, direct + transitive)'
    # NuGetAudit already fails restore on high/critical advisories; this stage
    # additionally reports EVERY known-vulnerable package so lower severities
    # are visible in the gate output, and fails on any finding.
    $output = @(dotnet list package --vulnerable --include-transitive 2>&1)
    Assert-LastExitCode 'audit (dotnet list package)'
    $output | ForEach-Object { Write-Host $_ }
    $findings = @($output | Where-Object { $_ -match 'has the following vulnerable packages' })
    if ($findings.Count -gt 0) {
        Write-Host 'FAIL: audit found vulnerable packages (see table above)' -ForegroundColor Red
        exit 1
    }
}

# Stage plan: a requested stage runs after its prerequisites so flags like
# --no-restore / --no-build are always valid.
$plan = switch ($Stage) {
    'all'     { @('restore', 'format', 'build', 'test', 'audit') }
    'restore' { @('restore') }
    'format'  { @('restore', 'format') }
    'build'   { @('restore', 'build') }
    'test'    { @('restore', 'build', 'test') }
    'audit'   { @('restore', 'audit') }
}

foreach ($step in $plan) {
    switch ($step) {
        'restore' { Invoke-RestoreStage }
        'format'  { Invoke-FormatStage }
        'build'   { Invoke-BuildStage }
        'test'    { Invoke-TestStage }
        'audit'   { Invoke-AuditStage }
    }
}

Write-Host ''
Write-Host 'verify: OK' -ForegroundColor Green
exit 0
