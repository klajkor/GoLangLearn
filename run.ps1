<#
.SYNOPSIS
  Build and run GoLearn.
.EXAMPLE
  ./run.ps1          # build frontend, start server on http://127.0.0.1:8080 and open the browser
  ./run.ps1 dev      # Go server + Vite dev server with hot reload on http://localhost:5173
  ./run.ps1 build    # produce a single golearn.exe
  ./run.ps1 test     # run Go tests and frontend type-check
  ./run.ps1 reset    # erase all learning progress (asks first); add -Force to skip the question
#>
param(
    [ValidateSet('run', 'dev', 'build', 'test', 'reset')][string]$Mode = 'run',
    [switch]$Force
)

$ErrorActionPreference = 'Stop'
Set-Location $PSScriptRoot

# winget installs Go here; a terminal opened before the install won't have it on PATH yet.
if (-not (Get-Command go -ErrorAction SilentlyContinue) -and (Test-Path 'C:\Program Files\Go\bin\go.exe')) {
    $env:PATH = "C:\Program Files\Go\bin;$env:PATH"
}

function Build-Web {
    Push-Location web
    try {
        if (-not (Test-Path node_modules)) { npm install }
        npm run build
        if ($LASTEXITCODE -ne 0) { throw 'frontend build failed' }
    } finally { Pop-Location }
}

switch ($Mode) {
    'run' {
        Build-Web
        go run ./cmd/server -open
    }
    'dev' {
        if (-not (Test-Path web/node_modules)) { Push-Location web; npm install; Pop-Location }
        $server = Start-Process go -ArgumentList 'run', './cmd/server' -NoNewWindow -PassThru
        try {
            Push-Location web
            npm run dev -- --open
        } finally {
            Pop-Location
            Stop-Process -Id $server.Id -ErrorAction SilentlyContinue
            Get-Process server -ErrorAction SilentlyContinue | Stop-Process -ErrorAction SilentlyContinue
        }
    }
    'build' {
        Build-Web
        go build -o golearn.exe ./cmd/server
        Write-Host 'Built golearn.exe: one self-contained file (UI and course included). Run: .\golearn.exe -open'
    }
    'reset' {
        $files = 'data/progress.db', 'data/progress.db-wal', 'data/progress.db-shm' | Where-Object { Test-Path $_ }
        if (-not $files) {
            Write-Host 'Nothing to reset: there is no progress database yet.'
            return
        }
        if (Get-Process -Name server, golearn -ErrorAction SilentlyContinue) {
            Write-Warning 'GoLearn looks like it is running. Stop it (Ctrl+C in its terminal) before resetting.'
        }
        if (-not $Force) {
            $answer = Read-Host 'This permanently erases all XP, levels, streaks, badges, completed lessons and quiz results. Type yes to continue'
            if ($answer -ne 'yes') {
                Write-Host 'Cancelled. Nothing was changed.'
                return
            }
        }
        Remove-Item $files -Force
        Write-Host 'Progress erased. The next start begins from scratch.'
    }
    'test' {
        go vet ./...
        go test ./...
        Push-Location web; npx tsc --noEmit; Pop-Location
    }
}
