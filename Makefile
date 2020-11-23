test:
	docker run --rm -v `pwd`:/go/src/github.com/gokits/cfg golang:1.15.5 /bin/bash -c 'cd /go/src/github.com/gokits/cfg; GOPROXY=https://goproxy.cn,direct GO111MODULE=on go test ./...'
