@echo off
setlocal

echo === keygen-cli local deployer ===
echo.

REM --- Build ---
echo [1/3] Building keygen.exe ...
go build -buildvcs=false -ldflags "-s -w -X main.version=0.1.0" -o keygen.exe .
if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed.
    exit /b 1
)
echo       OK

REM --- Copy binary ---
set "BIN_TARGET=c:\data\dev\keygen.exe"
echo [2/3] Copying keygen.exe to %BIN_TARGET% ...
copy /Y keygen.exe "%BIN_TARGET%" >nul
if %ERRORLEVEL% neq 0 (
    echo ERROR: Could not copy binary.
    exit /b 1
)
echo       OK

REM --- Copy .env ---
set "ENV_DIR=%USERPROFILE%\.keygen-cli"
set "ENV_TARGET=%ENV_DIR%\.env"
echo [3/3] Copying .env to %ENV_TARGET% ...
if not exist "%ENV_DIR%" mkdir "%ENV_DIR%"
if exist ".env" (
    copy /Y .env "%ENV_TARGET%" >nul
    if %ERRORLEVEL% neq 0 (
        echo ERROR: Could not copy .env file.
        exit /b 1
    )
    echo       OK
) else (
    echo       SKIPPED ^(.env not found in project root^)
)

echo.
echo Deploy complete. Run "keygen --version" from anywhere.
endlocal
