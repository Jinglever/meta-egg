# PowerShell 脚本：install.ps1
$ErrorActionPreference = "Stop"

$repo = "Jinglever/meta-egg"
$binary = "meta-egg.exe"
$installDir = "$HOME\.meta-egg\bin"

# 检测架构
$arch = if ([System.Environment]::Is64BitOperatingSystem) { "amd64" } else { "arm64" }

# 检测操作系统
$os = "windows"

# 获取最新版本号
$latest = (Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest").tag_name
if (-not $latest) {
    Write-Host "Failed to fetch latest version." -ForegroundColor Red
    exit 1
}

# 组装下载链接
$zipfile = "meta-egg-$os-$arch.zip"
$url = "https://github.com/$repo/releases/download/$latest/$zipfile"

# 创建安装目录
if (-not (Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir | Out-Null
}

# 下载并解压
$tmp = New-TemporaryFile
$tmpZip = "$($tmp.FullName).zip"
Write-Host "Downloading $url ..."
Invoke-WebRequest -Uri $url -OutFile $tmpZip

Write-Host "Extracting $zipfile ..."
Add-Type -AssemblyName System.IO.Compression.FileSystem
[System.IO.Compression.ZipFile]::ExtractToDirectory($tmpZip, $installDir)

# 添加到用户 PATH（如果未添加）
$envPath = [System.Environment]::GetEnvironmentVariable("PATH", "User")
if ($envPath -notlike "*$installDir*") {
    Write-Host "Adding $installDir to your user PATH..."
    [System.Environment]::SetEnvironmentVariable("PATH", "$envPath;$installDir", "User")
    Write-Host "You may need to restart your terminal for PATH changes to take effect." -ForegroundColor Yellow
}

Remove-Item $tmpZip

Write-Host "meta-egg ($latest) installed successfully!" -ForegroundColor Green
Write-Host "Run 'meta-egg --help' to get started."

# 生成 PowerShell completion
$completionScript = "$HOME\meta-egg-completion.ps1"
& "$installDir\$binary" completion powershell > $completionScript
Write-Host "PowerShell completion script generated at $completionScript"

# 自动添加到 $PROFILE
if (-not (Test-Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force | Out-Null
}
if (-not (Select-String -Path $PROFILE -Pattern "meta-egg-completion.ps1" -Quiet)) {
    Add-Content -Path $PROFILE -Value ". '$completionScript'"
    Write-Host "Added completion script to your PowerShell profile ($PROFILE)."
    Write-Host "Restart PowerShell to enable meta-egg completion."
} else {
    Write-Host "Completion script already referenced in your PowerShell profile."
} 