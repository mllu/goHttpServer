GO=	env GOPATH=`pwd` go

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

PKG = pkg/$(GOOS)_$(GOARCH)

ALL = goHttpServer

all: .gpm $(ALL)

.gpm: Godeps
	env GOPATH=`pwd` gpm
	touch $@

.phony: .gpm

REPOS = 		\
				src/git.apache.org \
				src/github.com \
				src/golang.org \
				src/gopkg.in \
				$(NULL)

goHttpServer: \
	$(PKG)/dogstats.a \
	$(NULL)
	$(GO) build goHttpServer
	$(GO) install goHttpServer

$(PKG)/dogstats.a: \
	src/util/dogstats/stats.go \
	$(NULL)
	$(GO) build util/dogstats
	$(GO) install util/dogstats

clean: goclean
	-rm -f *.o *~

goclean:
	-rm -rf pkg $(REPOS)
	-rm -f .gpm
	-rm -f $(ALL)
	-rm -f bin/$(ALL)
