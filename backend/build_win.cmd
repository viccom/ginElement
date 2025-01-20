@echo off
REM 项目名称
set PROJECT_NAME=goiot

REM Go 编译器
set GO=go

REM 支持的架构
set ARCHS=386 amd64

REM 编译输出目录
set OUTPUT_DIR=winbuild

REM 清理编译输出目录
echo Cleaning build directory...
if exist %OUTPUT_DIR% (
    rmdir /s /q %OUTPUT_DIR%
)

REM 创建输出目录
mkdir %OUTPUT_DIR%

REM 遍历所有架构并编译
for %%A in (%ARCHS%) do (
    echo "Building for windows/%%A..."
    mkdir %OUTPUT_DIR%\windows_%%A
    set GOARCH=%%A
    set GOOS=windows
    set CGO_ENABLED=0
    %GO% build -ldflags="-s -w" -o %OUTPUT_DIR%\windows_%%A\%PROJECT_NAME%_win%GOARCH%.exe .
)

echo "Build completed!"
pause