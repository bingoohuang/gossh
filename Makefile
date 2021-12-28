.PHONY: default test
all: default test

APPNAME=gossh

default:
	gofmt -s -w .&&go mod tidy&&go fmt ./...&&revive .&&goimports -w .&&golangci-lint run&&go install -ldflags="-s -w" ./...

install:
	go install -ldflags="-s -w" ./...
	ls -lh ~/go/bin/$(APPNAME)*

installlinux:
	env GOOS=linux GOARCH=amd64 go install -ldflags="-s -w" ./...
	ls -lh ~/go/bin/linux_amd64/$(APPNAME)*

packagelinux:installlinux
	ls -lh ~/go/bin/linux_amd64/$(APPNAME)*
	upx ~/go/bin/linux_amd64/$(APPNAME)
	ls -lh ~/go/bin/linux_amd64/$(APPNAME)*
	mv ~/go/bin/linux_amd64/$(APPNAME) ~/go/bin/linux_amd64/$(APPNAME)-linux_amd64
	gzip -f ~/go/bin/linux_amd64/$(APPNAME)-linux_amd64
	ls -lh ~/go/bin/linux_amd64/$(APPNAME)*

package: install
	ls -lh ~/go/bin/$(APPNAME)*
	upx ~/go/bin/$(APPNAME)
	ls -lh ~/go/bin/$(APPNAME)*
	mv ~/go/bin/$(APPNAME) ~/go/bin/$(APPNAME)-darwin-amd64
	gzip -f ~/go/bin/$(APPNAME)-darwin-amd64
	ls -lh ~/go/bin/$(APPNAME)*
test:
	go test ./...

# https://hub.docker.com/_/golang
# docker run --rm -v "$PWD":/usr/src/myapp -v "$HOME/dockergo":/go -w /usr/src/myapp golang make docker
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang bash
# 静态连接 glibc
docker:
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$HOME/dockergo":/go -w /usr/src/myapp golang make dockerinstall
	upx ~/dockergo/bin/$(APPNAME)
	mv ~/dockergo/bin/$(APPNAME)  ~/dockergo/bin/$(APPNAME)-amd64-glibc2.28
	gzip ~/dockergo/bin/$(APPNAME)-amd64-glibc2.28

dockerinstall:
	go install -v -x -a -ldflags '-extldflags "-static" -s -w' ./...
