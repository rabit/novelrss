CMD=novelrss
GO_SRCS=$(shell find -type f -name "*.go" -not -path "./vendor/*")
ASSETS_SRCS=$(shell find web/ -type f)

$(CMD): packr2 $(GO_SRCS) novelrss-packr.go
	go build -o $(CMD) cmd/main.go

novelrss-packr.go: $(ASSETS_SRCS)
	packr2 build

.PHONY: packr2
packr2: $(GOPATH)/bin/packr2

$(GOPATH)/bin/packr2:
	go get -u github.com/gobuffalo/packr/v2/packr2

.PHONY: clean
clean:
	packr2 clean
	rm -f $(CMD)

.PHONY: variables
variables:
	@echo CMD=$(CMD)
	@echo GO_SRCS=$(GO_SRCS)
	@echo ASSETS_SRCS=$(ASSETS_SRCS)

.PHONY: distclean
distclean: clean
	sudo rm -rf vendor/pkg
	rm -rf vendor/bin
	rm -rf vendor/src
