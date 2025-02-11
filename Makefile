.SILENT:
.PHONY: build build-linux build-windows pack

APP:=categraf
ROOT:=$(shell pwd -P)
GIT_COMMIT:=$(shell git --work-tree ${ROOT}  rev-parse 'HEAD^{commit}')
_GIT_VERSION:=$(shell git --work-tree ${ROOT} describe --tags --abbrev=14 "${GIT_COMMIT}^{commit}" 2>/dev/null)
TAG=$(shell echo "${_GIT_VERSION}" |  awk -F"-" '{print $$1}')
# GIT_VERSION:="$(TAG)-$(GIT_COMMIT)"
GIT_VERSION:="$(TAG)"
BUILD_VERSION:='flashcat.cloud/categraf/config.Version=$(GIT_VERSION)'
LDFLAGS:="-w -s -X $(BUILD_VERSION)"

all: build

vendor:
	GOPROXY=https://goproxy.cn go mod vendor

build:
	echo "Building version $(GIT_VERSION)"
	go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP)

build-pure:
	echo "Building version $(GIT_VERSION)"
	go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP)

build-linux:
	echo "Building version $(GIT_VERSION) for linux"
	GOOS=linux GOARCH=amd64 go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP)

build-linux-arm:
	echo "Building version $(GIT_VERSION) for linux"
	GOOS=linux GOARCH=arm64 go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP)

build-windows:
	echo "Building version $(GIT_VERSION) for windows"
	GOOS=windows GOARCH=amd64 go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP).exe

build-mac:
	echo "Building version $(GIT_VERSION) for mac"
	GOOS=darwin GOARCH=amd64 go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP).mac

build-mac-arm:
	echo "Building version $(GIT_VERSION) for mac"
	GOOS=darwin GOARCH=arm64 go build --tags "no_prometheus no_traces no_ibex no_logs" -ldflags $(LDFLAGS) -o $(APP).mac

build-image: build-linux
	echo "Building image flashcatcloud/categraf:$(TAG)"
	cp -rf conf/config.toml conf/logs.toml conf/input.cpu conf/input.disk conf/input.diskio conf/input.docker conf/input.kernel conf/input.kernel_vmstat conf/input.linux_sysctl_fs conf/input.mem conf/input.net conf/input.netstat conf/input.processes conf/input.system docker/conf/ && cp -f categraf docker/ && cd docker && docker build -t flashcatcloud/categraf:$(TAG) .

pack:build-linux build-windows
	rm -rf $(APP)-$(TAG).tar.gz
	rm -rf $(APP)-$(TAG).zip
	tar -zcvf $(APP)-$(TAG)-linux-amd64.tar.gz conf $(APP)
	zip -r $(APP)-$(TAG)-windows-amd64.zip conf $(APP).exe

release: build-linux build-windows
	rm -rf release/
	mkdir -p release/conf
	cp -rf conf.default/* release/conf
	mv release/conf/install.bat release
	mv release/conf/install.sh release
	cp $(APP) release
	cp $(APP).exe release
	tar -zcvf release/$(APP)-$(TAG)-linux-amd64.tar.gz -C release conf install.sh $(APP)
	cd release && zip -r $(APP)-$(TAG)-windows-amd64.zip conf install.bat $(APP).exe

version: build-linux build-windows
	mv $(APP) $(APP)-$(TAG)-linux-amd64
	gzip -f $(APP)-$(TAG)-linux-amd64
	mv $(APP).exe $(APP)-$(TAG)-windows-amd64.exe
	gzip -f $(APP)-$(TAG)-windows-amd64.exe