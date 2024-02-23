staticcheck ./...
@if %ERRORLEVEL% neq 0 exit /b %ERRORLEVEL%
go vet ./...
@if %ERRORLEVEL% neq 0 exit /b %ERRORLEVEL%
go test ./...
@if %ERRORLEVEL% neq 0 exit /b %ERRORLEVEL%
mkdir build
set "GOOS=linux"   & go build -o ./build/ ./...
set "GOOS=windows" & go build -o ./build/ ./...
@if %ERRORLEVEL% neq 0 exit /b %ERRORLEVEL%
