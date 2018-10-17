all: build
.PHONY: all

build:
	go build github.com/openshift/library-go/pkg/...
.PHONY: build

test:
	go test github.com/openshift/library-go/pkg/...
.PHONY: test

verify:
	hack/verify-golint.sh
.PHONY: verify

update-deps:
	hack/update-deps.sh
.PHONY: update-deps
