SHELL := /bin/bash
DOCKER_IMAGE_NAME := biblia2y
GIT_COMMIT?=$(shell git rev-parse HEAD 2>/dev/null)
GIT_BRANCH?=$(shell git rev-parse --abbrev-ref HEAD | sed  "s/\//_/g" 2>/dev/null)
DOCKER_REGISTRY?=518205865037.dkr.ecr.us-east-1.amazonaws.com

.PHONY: auth
auth:
	aws ecr get-login --no-include-email --profile private | bash

.PHONY: bin
bin:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o bin/server cmd/main.go
	
.PHONY: build
build: bin
	mkdir ./db-empty-dir
	docker build --no-cache -f Dockerfile -t $(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_$(GIT_COMMIT) .
	rm -r ./db-empty-dir

.PHONY: push
push: build
	docker tag $(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_$(GIT_COMMIT) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_$(GIT_COMMIT)
	docker tag $(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_$(GIT_COMMIT) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_latest
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_$(GIT_COMMIT)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE_NAME)\:$(GIT_BRANCH)_latest


backup:
	scp -r biblia2y.qradium.com:db backup

.PHONY: test
test:
	go test ./...
