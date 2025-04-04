# 构建指南

本文档提供了如何在不同平台上编译 clash-speedtest 的说明。

## 前提条件

- 安装 [Go](https://golang.org/dl/) 1.23 或更高版本
- 确保已正确设置 GOPATH 和 GOROOT 环境变量

## 使用构建脚本

### Windows 平台

在 Windows 上，你可以使用提供的批处理脚本来编译项目：

```cmd
.\build.bat
```

这将在 `dist` 目录下生成以下文件：
- `clash-speedtest-windows-amd64.exe`：Windows 64位版本
- `clash-speedtest-windows-arm64.exe`：Windows ARM64版本
- `clash-speedtest-linux-amd64`：Linux 64位版本
- `clash-speedtest-linux-arm64`：Linux ARM64版本
- `clash-speedtest-darwin-amd64`：macOS 64位版本
- `clash-speedtest-darwin-arm64`：macOS ARM64版本

### Linux/macOS 平台

在 Linux 或 macOS 上，你可以使用提供的 shell 脚本来编译项目：

```bash
# 确保脚本有执行权限
chmod +x ./build.sh

# 执行构建脚本
./build.sh
```

这将在 `dist` 目录下生成与 Windows 平台相同的输出文件。

## 使用 Makefile

如果你的系统安装了 `make` 工具，你可以使用以下命令：

```bash
# 编译所有平台版本
make all

# 仅编译 Linux 版本
make linux

# 仅编译 Windows 版本
make windows

# 仅编译 macOS 版本
make darwin

# 清理构建目录
make clean
```

## 手动编译

如果你想手动编译特定平台的版本，可以使用以下命令：

### Linux

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o clash-speedtest .
```

### Windows

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o clash-speedtest.exe .
```

### macOS

```bash
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o clash-speedtest .
```

## 注意事项

- 编译时设置 `CGO_ENABLED=0` 可以生成静态链接的二进制文件，提高可移植性
- 如果需要减小二进制文件大小，可以添加 `-ldflags="-s -w"` 参数
- 编译后的二进制文件可以直接运行，无需额外依赖