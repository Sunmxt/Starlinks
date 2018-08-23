.PHONY: dep-gin clean dep-gin env

export GOPATH:=$(shell pwd)
VENDOR_BIN:=$(shell which govendor 2>/dev/null)

bin/starlinks: dep-gin
	echo Not avaliable.

dep-gin: bin/govendor
	pushd $(GOPATH)/src; \
	../bin/govendor fetch github.com/gin-gonic/gin@v1.3; \
	popd

vendor-init:
	pushd $(GOPATH)/src; \
	../bin/govendor init; \
	popd 

bin/govendor: 
ifeq ($(VENDOR_BIN),)
	go get -u github.com/kardianos/govendor
	rm -rf $(GOPATH)/src/github.com/kardianos/govendor
else
	ln -s $(VENDOR_BIN) bin/govendor
endif

clean:
	rm bin/govendor
