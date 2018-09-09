.PHONY: dep-gin clean dep-gin env release debug fmt list-deps new-dep

export GOPATH:=$(shell pwd)
VENDOR_BIN:=$(shell which govendor 2>/dev/null)

debug: dep fmt
	go install -gcflags '-N -l' starlinks/starlinks

release: dep fmt
	go install -ldflags "-s" starlinks/starlinks

fmt:
	go fmt starlinks/...

dep: bin/govendor
	cd src; \
	../bin/govendor sync; \

vendor-init:
	cd src; \
	../bin/govendor init; \

bin/govendor: 
ifeq ($(VENDOR_BIN),)
	go get -u github.com/kardianos/govendor
	rm -rf $(GOPATH)/src/github.com/kardianos/govendor
else
	ln -s $(VENDOR_BIN) bin/govendor
endif

list-deps: bin/govendor
	cd src;\
	../bin/govendor list;\

new-dep: bin/govendor
ifeq ($(PACKAGE),)
	@echo Add new dependency
	@echo Usage: make new-deps PACKAGE='<path of new package>'
else
	@echo bin/govendor fetch $(PACKAGE)
	cd src;\
	../bin/govendor fetch $(PACKAGE);
endif

clean:
	rm bin/govendor bin/starlinks
