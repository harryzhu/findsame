rm -rf ./dist/*

CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o dist/macos_arm/findsame -ldflags "-w -s" main.go
zip dist/macos_arm/findsame_macos_arm.zip dist/macos_arm/findsame

CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o dist/macos_intel/findsame -ldflags "-w -s" main.go
zip dist/macos_intel/findsame_macos_intel.zip dist/macos_intel/findsame


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/linux_amd64/findsame -ldflags "-w -s" main.go
zip dist/linux_amd64/findsame_linux_amd64.zip dist/linux_amd64/findsame


CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/windows_amd64/findsame.exe -ldflags "-w -s" main.go
zip dist/windows_amd64/findsame_windows_amd64.zip dist/windows_amd64/findsame.exe
