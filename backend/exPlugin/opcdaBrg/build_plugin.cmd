@echo off
cd /d "%~dp0"
REM 项目名称
set PROJECT_NAME=opcdaBrg

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

REM 启用延迟变量扩展
setlocal enabledelayedexpansion

REM 遍历所有架构并编译
for %%A in (%ARCHS%) do (
    set GOARCH=%%A
    set GOOS=windows
    set CGO_ENABLED=0
    echo Building for windows !GOARCH!...
    %GO% build -ldflags="-s -w" -o %OUTPUT_DIR%\%PROJECT_NAME%_win!GOARCH!.exe .
    upx -9 %OUTPUT_DIR%\%PROJECT_NAME%_win!GOARCH!.exe
)

copy brokerAuth.json %OUTPUT_DIR%\
echo Build completed!
pause