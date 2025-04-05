# 设置交叉编译的环境变量
$env:GOOS = "linux"
$env:GOARCH = "arm"
$env:GOARM = "7"
$env:CGO_ENABLED = "0"

# 创建发布目录
$releaseDir = "release"
if (!(Test-Path $releaseDir)) {
    New-Item -ItemType Directory -Path $releaseDir
} else {
    # 清理旧文件
    Remove-Item "$releaseDir/*" -Force -Recurse
}

Write-Host "正在编译 ARM 版本..."

# 编译到 release 目录
go build -o "$releaseDir/aku-web"

if ($LASTEXITCODE -eq 0) {
    Write-Host "ARM 版本编译成功！"
} else {
    Write-Host "ARM 版本编译失败！"
    exit 1
}

# 打包静态文件
Write-Host "正在打包静态文件..."
Push-Location
Set-Location $releaseDir
tar -cf static.tar ../static
Pop-Location

Write-Host "编译和打包完成！文件已保存到 release 目录："
Get-ChildItem $releaseDir | ForEach-Object {
    Write-Host "- $($_.Name)"
}

