GOMODCACHE ?= ${HOME}/.gomodcache
GOPROXY ?= "https://goproxy.cn,direct"
DOCKERGO ?=	docker run --rm \
		-e GOMODCACHE=/tmp/gomodcache \
		-e GOCACHE=/tmp/gocache \
		-e GOPROXY=${GOPROXY} \
		-e GO111MODULE=on \
		--user=${shell id -u}:${shell id -g} \
		-v ${GOMODCACHE}:/tmp/gomodcache:rw \
		-v `pwd`:/tmp/proj:rw \
		-w /tmp/proj \
		golang:1.15.3 

test:
	${DOCKERGO} go test ./...

fmt:
	${DOCKERGO} go fmt
