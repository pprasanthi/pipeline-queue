MAJOR_MINOR = 0.2
BUILD = $(shell  date -u "+%Y%m%d-%H%M%S")
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null || echo nocommitinfo)

export CGO_ENABLED ?= 0
export VERSION 	= $(MAJOR_MINOR).$(BUILD).$(COMMIT_HASH)
export FLAGS = $(shell echo "\
        -X gitlab.com/pprasanthi/job-queue/internal.buildBranch=$(shell git rev-parse --abbrev-ref HEAD) \
        -X gitlab.com/pprasanthi/job-queue/internal.buildCompiler=$(shell go version | cut -f 3 -d' ') \
        -X gitlab.com/pprasanthi/job-queue/internal.buildDate=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ') \
        -X gitlab.com/pprasanthi/job-queue/internal.buildHash=$(COMMIT_HASH) \
        -X gitlab.com/pprasanthi/job-queue/internal.buildUser=$(USER) \
        -X gitlab.com/pprasanthi/job-queue/internal.buildVersion=$(VERSION)")
export SRC=/go/src/gitlab.com/pprasanthi/job-queue

clean:
	rm -rf target vendor

build:
	@./cicd/build.sh

gitlab-init:
	mkdir -p $(CI_PROJECT_DIR)/artifacts
	mkdir -p /go/src/gitlab.com/pprasanthi
	cp -r $(CI_PROJECT_DIR) $(SRC)

gitlab-archive:
	cp -r $(SRC)/target/* $(CI_PROJECT_DIR)/artifacts/

install:
	go get -u github.com/golang/dep/cmd/dep

vendor:
	dep ensure

docker-build:
	docker build --target Builder -t pprasanthi/job-queue .

docker:
	docker build -t pprasanthi/job-queue .
	docker tag pprasanthi/job-queue:latest pprasanthi/singleton-job-queue:$(MAJOR_MINOR)

docker-publish: docker
	@./cicd/docker_publish.sh pprasanthi/singleton-job-queue

test:
	go test -v ./...
