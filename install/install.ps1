param(
    [string]$Version = "latest",
    [string]$Repo = "msherburne/dokbox",
    [string]$BaseUrl = "",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

function Show-Usage {
    @"
Usage: install.ps1 [-Version <version>] [-Repo <owner/repo>] [-BaseUrl <url>] [-Help]

Install Dokbox using a local winget manifest generated from GitHub release assets.
"@
}

function Resolve-Tag {
    param([string]$InputVersion)

    if ($InputVersion -eq "latest") {
        return "latest"
    }

    if ($InputVersion.StartsWith("v")) {
        return $InputVersion
    }

    return "v$InputVersion"
}

function Resolve-BaseUrl {
    param(
        [string]$InputRepo,
        [string]$InputVersion,
        [string]$InputBaseUrl
    )

    if ($InputBaseUrl) {
        return $InputBaseUrl
    }

    if ($InputVersion -eq "latest") {
        return "https://github.com/$InputRepo/releases/latest/download"
    }

    return "https://github.com/$InputRepo/releases/download/$(Resolve-Tag -InputVersion $InputVersion)"
}

if ($Help) {
    Show-Usage
    exit 0
}

if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    throw "winget is required to install Dokbox on Windows."
}

$baseUrl = Resolve-BaseUrl -InputRepo $Repo -InputVersion $Version -InputBaseUrl $BaseUrl
$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("dokbox-install-" + [System.Guid]::NewGuid().ToString("N"))
$manifestZip = Join-Path $tempDir "dokbox-winget-manifests.zip"
$manifestDir = Join-Path $tempDir "winget"

New-Item -ItemType Directory -Path $manifestDir -Force | Out-Null

try {
    Invoke-WebRequest -Uri "$baseUrl/dokbox-winget-manifests.zip" -OutFile $manifestZip
    Expand-Archive -Path $manifestZip -DestinationPath $manifestDir -Force
    winget install --manifest $manifestDir --accept-package-agreements --accept-source-agreements
    Write-Host "Dokbox installation completed."
}
finally {
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force
    }
}
