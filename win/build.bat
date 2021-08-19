rem rsrc.exe -manifest main.manifest -o main.syso -ico main.ico
go build -ldflags "-s -w -H windowsgui" -o cos-transmit.exe
rem go build -ldflags "-s -w" -o cos-transmit.exe
rem go build
rem upx cos-transmit.exe