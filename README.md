# Categraf
[![Release](https://github.com/flashcatcloud/categraf/workflows/Release/badge.svg)](https://github.com/flashcatcloud/categraf/actions?query=workflow%3ARelease)
[![Powered By Flashcat](https://img.shields.io/badge/Powered%20By-Flashcat-red)](https://flashcat.cloud/)

Categraf is a monitoring agent for nightingale / prometheus / m3db / victoriametrics / thanos / influxdb / tdengine.

[![dockeri.co](https://dockeri.co/image/flashcatcloud/categraf)](https://hub.docker.com/r/flashcatcloud/categraf/)

## Links

- [QuickStart](https://www.gitlink.org.cn/flashcat/categraf/wiki/QuickStart)
- [FAQ](https://www.gitlink.org.cn/flashcat/categraf/wiki/FAQ)
- [Github Releases](https://github.com/flashcatcloud/categraf/releases)
- [Gitlink Releases](https://www.gitlink.org.cn/flashcat/categraf/releases)

## Build

```shell
# export GO111MODULE=on
# export GOPROXY=https://goproxy.cn
go build
```

## Pack

```shell
tar zcvf categraf.tar.gz categraf conf
```


## Run

```shell
# test mode: just print metrics to stdout
./categraf --test

# test system and mem plugins
./categraf --test --inputs system:mem

# print usage message
./categraf --help

# run
./categraf

# run with specified config directory
./categraf --configs /path/to/conf-directory

# only enable system and mem plugins
./categraf --inputs system:mem

# use nohup to start categraf
nohup ./categraf &> stdout.log &
```


## Deploy categraf as daemonset

edit k8s/daemonset.yaml, replace NSERVER_SERVICE_WITH_PORT with service ip:port of nserver in your cluster, replace CATEGRAF_NAMESPACE with namespace value, then run:

```shell
kubectl apply -n monitoring -f k8s/daemonset.yaml
kubectl apply -n monitoring -f k8s/sidecar.yaml
```
Notice: k8s/sidecar.yaml is a demo, replace mock with your own image.


## Plugin

plugin list: [https://github.com/flashcatcloud/categraf/tree/main/inputs](https://github.com/flashcatcloud/categraf/tree/main/inputs)


## Thanks

Categraf is developed on the basis of Telegraf, Exporters and the OpenTelemetry. Thanks to the great open source community.

## Community

![](doc/laqun.jpeg)
