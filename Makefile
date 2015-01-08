.PHONY: build install example test cover cover-web testlint lint

build:
	go build -race

install:
	go install -race .

example: install
	./example/test.sh

test:
	go test -race -cover -v -tags testing ./...

cover:
	t=`mktemp 2>/dev/null || mktemp -t 'cover'` && \
	go test -v -tags testing -race -covermode=set -coverprofile=$$t ./... ; \
	go tool cover -func=$$t ; \
	unlink $$t

cover-web:
	t=`mktemp 2>/dev/null || mktemp -t 'cover'` && \
	go test -v -tags testing -race -covermode=set -coverprofile=$$t ./... ; \
	go tool cover -html=$$t ; \
	unlink $$t

testlint:
	fgt gometalinter .

lint:
	gometalinter .
