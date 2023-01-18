SHELL = bash

BUILD_DIR              = build
DEVEL_BUILD_DIR        = $(BUILD_DIR)/devel
PROD_BUILD_DIR         = $(BUILD_DIR)/prod

GO_SRC_DIR             = .
GO_SRC_FILES           = $(wildcard $(GO_SRC_DIR)/*.go)
GO_DEVEL_BUILD_DIR     = $(DEVEL_BUILD_DIR)
GO_DEVEL_BUILD_BINARY  = $(GO_DEVEL_BUILD_DIR)/commento
GO_PROD_BUILD_DIR      = $(PROD_BUILD_DIR)
GO_PROD_BUILD_BINARY   = $(GO_PROD_BUILD_DIR)/commento

devel: devel-go

prod: prod-go

test: test-go

clean:
	rm -rf $(BUILD_DIR)

# There's really no difference between the prod and devel binaries in Go, but
# for consistency sake, we'll use separate targets (maybe this will be useful
# later down the line).

devel-go:
	go mod tidy
    # Statically link the executable so that we can run that in Alpine
	# -a forces a rebuild of all packages, which is what we want when we build against Alpine (using static linking)
	CGO_ENABLED=0 go build -a -v -o $(GO_DEVEL_BUILD_BINARY) -ldflags "-X main.version=$(shell git describe --tags)"

prod-go:
	go mod tidy
    # Statically link the executable so that we can run that in Alpine
	# -a forces a rebuild of all packages, which is what we want when we build against Alpine (using static linking)
	CGO_ENABLED=0 go build -a -v -o $(GO_PROD_BUILD_BINARY) -ldflags "-w -s -X main.version=$(shell git describe --tags)"

test-go:
	go mod tidy
	go test -v .

$(shell mkdir -p $(GO_DEVEL_BUILD_DIR) $(GO_PROD_BUILD_DIR))
