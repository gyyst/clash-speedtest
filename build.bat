@echo off

:: 清理旧的构建文件
if exist .\dist rmdir /s /q .\dist

:: 创建输出目录
mkdir .\dist

:: 编译Linux版本
echo 编译Linux版本...
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-linux-amd64 .

set GOOS=linux
set GOARCH=arm64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-linux-arm64 .

:: 编译Windows版本
echo 编译Windows版本...
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-windows-amd64.exe .

set GOOS=windows
set GOARCH=arm64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-windows-arm64.exe .

:: 编译macOS版本
echo 编译macOS版本...
set GOOS=darwin
set GOARCH=amd64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-darwin-amd64 .

set GOOS=darwin
set GOARCH=arm64
set CGO_ENABLED=0
go build -o .\dist\clash-speedtest-darwin-arm64 .

echo 编译完成，输出文件在 .\dist 目录下
dir .\dist