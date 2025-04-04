# Clash-SpeedTest Makefile

.PHONY: all clean linux windows darwin

all: linux windows darwin

clean:
	rm -rf ./dist
	if exist .\dist rmdir /s /q .\dist
	mkdir -p ./dist
	mkdir .\dist 2>nul || echo dist directory exists

linux: clean
	@echo "编译Linux版本..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-linux-amd64 .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-linux-arm64 .

windows: clean
	@echo "编译Windows版本..."
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-windows-arm64.exe .

darwin: clean
	@echo "编译macOS版本..."
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o ./dist/clash-speedtest-darwin-arm64 .