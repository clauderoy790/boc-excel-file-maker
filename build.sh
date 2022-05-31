# build for windows
GOOS=windows GOARCH=amd64 go build  -o bin/rates.exe -ldflags -H=windowsgui