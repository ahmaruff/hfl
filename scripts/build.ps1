# scripts/build.ps1
Write-Host "Building HFL binaries..."

$platforms = @(
    @{os="windows"; arch="amd64"; ext=".exe"},
    @{os="linux"; arch="amd64"; ext=""},
    @{os="darwin"; arch="amd64"; ext=""},
    @{os="darwin"; arch="arm64"; ext=""}  # M1 Macs
)

# Create dist directory
New-Item -ItemType Directory -Force -Path "dist"

foreach ($platform in $platforms) {
    $env:GOOS = $platform.os
    $env:GOARCH = $platform.arch
    $output = "dist/hfl-$($platform.os)-$($platform.arch)$($platform.ext)"
    
    Write-Host "Building $output..."
    go build -ldflags "-s -w" -o $output
}

Write-Host "Build complete! Check dist/ directory"
