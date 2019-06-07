.PHONY: build install

build:
	CGO_ENABLED=0 go build \
        -gcflags=-trimpath="/go/src/github.com/gradecak/benchmark" \
        -asmflags=-trimpath="/go/src/github.com/gradecak/benchmark" \
        -ldflags "-X \"main.buildTime=$(date)\" " \
		 $(CURDIR)/cmd/benchmark
