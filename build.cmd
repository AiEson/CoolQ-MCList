@echo off

echo Generating app.json
go build github.com\Tnze\CoolQ-Golang-SDK\tools\cqcfg
go generate
cqcfg C:\Users\23684\Downloads\Compressed\CoolQMCList-master
IF ERRORLEVEL 1 pause

echo Setting env vars
SET CGO_LDFLAGS=-Wl,--kill-at
SET CGO_ENABLED=1
SET GOOS=windows
SET GOARCH=386
SET GOPROXY=https://goproxy.cn

echo Building app.dll
go build -ldflags "-s -w" -buildmode=c-shared -o app.dll
IF ERRORLEVEL 1 pause