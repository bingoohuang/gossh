.PHONY: test install git.commit git.branch default
all: install

app=$(notdir $(shell pwd))
appVersion := 1.0.0
goVersion := $(shell go version | sed 's/go version //'|sed 's/ /_/')
# e.g. 2021-10-28T11:49:52+0800
buildTime := $(shell date +%FT%T%z)
# e.g. 20240808080013
buildTimeCompact := $(shell date +%Y%m%d%H%M%S)
# https://git-scm.com/docs/git-rev-list#Documentation/git-rev-list.txt-emaIem
# e.g. ffd23d3@2022-04-06T18:07:14+08:00
gitCommit := $(shell [ -f git.commit ] && cat git.commit || git log --format=format:'%h@%aI' -1)
gitBranch := $(shell [ -f git.branch ] && cat git.branch || git rev-parse --abbrev-ref HEAD)
gitInfo = $(gitBranch)-$(gitCommit)
#gitCommit := $(shell git rev-list -1 HEAD)
# https://stackoverflow.com/a/47510909
pkg := github.com/bingoohuang/gg/pkg/v

extldflags := -extldflags -static
# https://ms2008.github.io/2018/10/08/golang-build-version/
# https://github.com/kubermatic/kubeone/blob/master/Makefile
flags1 = -s -w -X $(pkg).BuildTime=$(buildTime) -X $(pkg).AppVersion=$(appVersion) -X $(pkg).GitCommit=$(gitInfo) -X $(pkg).GoVersion=$(goVersion)
flags2 = ${extldflags} ${flags1}
buildTags = $(if $(TAGS),-tags=$(TAGS),)
buildFlags = ${buildTags} -trimpath -ldflags="'${flags1}'"
goinstall_target = $(if $(TARGET),$(TARGET),./...)

goinstall = go install ${buildTags} -trimpath -ldflags='${flags1}' ${goinstall_target}
gobin := $(shell go env GOBIN)
# try $GOPATN/bin if $gobin is empty
gobin := $(if $(gobin),$(gobin),$(shell go env GOPATH)/bin)

export GOPROXY=https://mirrors.aliyun.com/goproxy/,https://goproxy.cn,https://goproxy.io,direct
# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

# usage: t=$(mktemp); echo $t; echo "set -x; go build -o build/rig_linux_arm64 $(make -f ~/github/gg/Makefile build.flags) ./cmd/rig" > $t && sh $t
build.flags:
	@echo ${buildFlags}

git.commit:
	echo ${gitCommit} > git.commit
	echo ${gitBranch} > git.branch

sec:
	@gosec ./...
	@echo "[OK] Go security check was completed!"

init:

lint-all:
	golangci-lint run --enable-all

lint:
	golangci-lint run ./...

fmt-update:
	go install mvdan.cc/gofumpt@latest
	go install golang.org/x/tools/cmd/...@latest 	# for goimports
	# go install github.com/mgechev/revive@master
	go install github.com/daixiang0/gci@latest
	go install github.com/google/osv-scanner/cmd/osv-scanner@latest
	go install github.com/polyfloyd/go-errorlint@latest
	go install github.com/dkorunic/betteralign/cmd/betteralign@latest
	go install -v github.com/go-critic/go-critic/cmd/gocritic@latest
	# Use right mirror functions for string/[]byte performance bust
	go install github.com/butuzov/mirror/cmd/mirror@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest

fmt:
	gofumpt -l -w .
	gofmt -s -w .
	go mod tidy
	go fmt ./...
	revive .
	goimports -w .
	gci write .
	osv-scanner -r .
	go-errorlint ./...
	gocritic check ./...
	betteralign ./...
	nilaway ./...
	# Use right mirror functions for string/[]byte performance bust
	# too slow
	# mirror ./...
	govulncheck ./...

align:
	betteralign -apply ./...

install-upx: init
	${goinstall}
	upx --best --lzma ${gobin}/${app}*
	ls -lh ${gobin}/${app}*

install: init
	${goinstall}
	ls -lh ${gobin}/${app}*

linux: init
	GOOS=linux GOARCH=amd64 ${goinstall}
	ls -lh  ${gobin}/linux_amd64/${app}*
linux-upx: init
	GOOS=linux GOARCH=amd64 ${goinstall}
	upx --best --lzma ${gobin}/linux_amd64/${app}*
	ls -lh  ${gobin}/linux_amd64/${app}*
windows: init
	GOOS=windows GOARCH=amd64 ${goinstall}
	ls -lh  ${gobin}/windows_amd64/${app}*
windows-upx: init
	GOOS=windows GOARCH=amd64 ${goinstall}
	upx --best --lzma ${gobin}/windows_amd64/${app}*
	ls -lh  ${gobin}/windows_amd64/${app}*
arm: init
	GOOS=linux GOARCH=arm64 ${goinstall}
	ls -lh  ${gobin}/linux_arm64/${app}*
arm-upx: init
	GOOS=linux GOARCH=arm64 ${goinstall}
	upx --best --lzma ${gobin}/linux_arm64/${app}*
	ls -lh  ${gobin}/linux_arm64/${app}*
mac-arm: init
	GOOS=darwin GOARCH=arm64 ${goinstall}
	# upx --best --lzma ${gobin}/darwin_arm64/${app}*
	ls -lh  ${gobin}/darwin_arm64/${app}*

upx:
	ls -lh ${gobin}/${app}*
	upx ${gobin}/${app}*
	ls -lh ${gobin}/${app}*
	ls -lh ${gobin}/linux_amd64/${app}*
	upx ${gobin}/linux_amd64/${app}*
	ls -lh ${gobin}/linux_amd64/${app}*

test: init
	#go test -v ./...
	go test -v -race ./...

bench: init
	#go test -bench . ./...
	go test -tags bench -benchmem -bench . ./...

clean:
	rm coverage.out

cover:
	go test -v -race -coverpkg=./... -coverprofile=coverage.out ./...

coverview:
	go tool cover -html=coverage.out

# https://hub.docker.com/_/golang
# docker run --rm -v "$PWD":/usr/src/myapp -v "$HOME/dockergo":/go -w /usr/src/myapp golang make docker
# docker run --rm -it -v "$PWD":/usr/src/myapp -w /usr/src/myapp golang bash
# 静态连接 glibc
docker:
	mkdir -p ~/dockergo
	docker run --rm -v "$$PWD":/usr/src/myapp -v "$$HOME/dockergo":/go -w /usr/src/myapp golang make dockerinstall
	#upx ~/dockergo/bin/${app}
	gzip -f ~/dockergo/bin/${app}

dockerinstall:
	go install -v -x -a -ldflags=${flags} ./...

targz: git.commit
	find . -name ".DS_Store" -delete
	#find . -type f -name '\.*' -print
	cd .. && tar czf ${app}.${buildTimeCompact}.tar.gz --exclude .git --exclude .idea  --no-xattrs --no-acls ${app} && ls -hl ${app}.${buildTimeCompact}.tar.gz

# BSSH_HOST=240f make bssh
bssh: targz
	bssh scp ../${app}.${buildTimeCompact}.tar.gz r:.
	bssh 'rm -fr ${app} && tar zxf ${app}.${buildTimeCompact}.tar.gz && cd ${app} && make install bin_cp && cd .. && mv ${app} ${app}.${buildTimeCompact} && cd ${app}.${buildTimeCompact} && ls -hl ./built/* && md5sum ./built/* && readlink -f ./built/*'
	mkdir -p ../${app}.${buildTimeCompact}.built
	bssh scp r:${app}.${buildTimeCompact}/built ./${app}.${buildTimeCompact}.built/
	# 显示大小
	ls -hl ./${app}.${buildTimeCompact}.built/built/*
	md5sum ./${app}.${buildTimeCompact}.built/built/*
	# 显示完整路径
	readlink -f ./${app}.${buildTimeCompact}.built/built/*

bin_cp:
	mkdir -p ./built && cp -r ${gobin}/${app}* ./built/
