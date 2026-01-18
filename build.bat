@echo off
echo Cleaning old files...
if exist multiablo_race.exe del /f /q multiablo_race.exe
if exist cmd\multiablo\rsrc_windows_amd64.syso del /f /q cmd\multiablo\rsrc_windows_amd64.syso

echo Building multiablo.exe...
echo Generating Windows resources...
go tool go-winres make --arch amd64 --in cmd/multiablo/winres/winres.json --out cmd/multiablo/rsrc

set CGO_ENABLED=1
go build -race -o multiablo_race.exe ./cmd/multiablo

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b %errorlevel%
)

echo Build successful! multiablo_race.exe created.
