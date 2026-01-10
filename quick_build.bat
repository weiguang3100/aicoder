@echo off
setlocal

set "APP_NAME=AICoder"
set "BIN_DIR=%~dp0build\bin"
set "GOPATH=%USERPROFILE%\go"
set "PATH=%GOPATH%\bin;%PATH%"
set "NSIS_PATH=C:\Program Files (x86)\NSIS\makensis.exe"

echo [1/5] Building frontend...
cd "%~dp0frontend"
call npm.cmd run build
if %errorlevel% neq 0 (
    echo [ERROR] Frontend build failed.
    exit /b 1
)
cd "%~dp0"

echo [2/5] Generating resource files...
"%GOPATH%\bin\rsrc.exe" -ico "%~dp0build\windows\icon.ico" -arch amd64 -o "%~dp0resource_windows_amd64.syso"
"%GOPATH%\bin\rsrc.exe" -ico "%~dp0build\windows\icon.ico" -arch arm64 -o "%~dp0resource_windows_arm64.syso"

echo [3/5] Building amd64...
set "GOOS=windows"
set "CGO_ENABLED=0"
set "GOARCH=amd64"
go build -tags desktop,production -ldflags "-s -w -H windowsgui" -o "%BIN_DIR%\%APP_NAME%_amd64.exe"

echo [4/5] Building arm64...
set "GOARCH=arm64"
go build -tags desktop,production -ldflags "-s -w -H windowsgui" -o "%BIN_DIR%\%APP_NAME%_arm64.exe"

echo [5/5] Creating NSIS installer...
"%NSIS_PATH%" /DINFO_PRODUCTVERSION="3.1.0.3100" /DARG_WAILS_AMD64_BINARY="%BIN_DIR%\%APP_NAME%_amd64.exe" /DARG_WAILS_ARM64_BINARY="%BIN_DIR%\%APP_NAME%_arm64.exe" "%~dp0build\windows\installer\multiarch.nsi"

del "%~dp0resource_windows_amd64.syso" 2>nul
del "%~dp0resource_windows_arm64.syso" 2>nul

echo [SUCCESS] Build complete!
endlocal
