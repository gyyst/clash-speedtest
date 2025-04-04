#!/bin/bash

# 清理旧的构建文件
rm -rf ./dist

# 创建输出目录
mkdir -p ./dist

# 编译Linux版本
echo "编译Linux版本..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-linux-amd64 .
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-linux-arm64 .

# 编译Windows版本
echo "编译Windows版本..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-windows-amd64.exe .
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-windows-arm64.exe .

# 编译macOS版本
echo "编译macOS版本..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-darwin-amd64 .
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-darwin-arm64 .

echo "编译完成，输出文件在 ./dist 目录下"
ls -la ./dist