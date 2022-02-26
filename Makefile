.PHONY: all windows linux amd64 386

all: windows linux

windows: ruflood-windows-386.exe ruflood-windows-amd64.exe

linux: ruflood-linux-386 ruflood-linux-amd64

amd64: ruflood-linux-amd64 ruflood-windows-amd64.exe

386: ruflood-linux-386 ruflood-windows-386.exe

clean:
	rm -f ruflood-linux-386 ruflood-linux-amd64 ruflood-windows-386.exe ruflood-windows-amd64.exe

ruflood-linux-386: go.mod go.sum ruflood.go
	env GOOS=linux GOARCH=386 go build -o ruflood-linux-386

ruflood-linux-amd64: go.mod go.sum ruflood.go
	env GOOS=linux GOARCH=amd64 go build -o ruflood-linux-amd64

ruflood-windows-386.exe: go.mod go.sum ruflood.go
	env GOOS=windows GOARCH=386 go build -o ruflood-windows-386.exe

ruflood-windows-amd64.exe: go.mod go.sum ruflood.go
	env GOOS=windows GOARCH=amd64 go build -o ruflood-windows-amd64.exe
