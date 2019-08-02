test:
	docker run --rm -v `pwd`:/go/src/github.com/gokits/cfg golang:1.12.7 /bin/bash -c 'cd /go/src/github.com/gokits/cfg; GOPROXY=https://goproxy.io GO111MODULE=on go test ./...'
