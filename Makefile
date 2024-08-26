SWIG_FILES:=$(shell find pkg -name '*.i')
GO_FILES:=$(SWIG_FILES:.i=_wrap.go)
SWIG_DIRS:=$(addsuffix cleandir,$(dir $(SWIG_FILES)))

build: prepare
	mkdir -p build
	go build -o build/device-controller

run: prepare
	go run .

prepare: swig

clean: $(SWIG_DIRS)
	@rm -rf build

%cleandir:
	@echo "remove $**_wrap.go and $**_wrap.cxx files"
	@rm $**_wrap.go $**_wrap.cxx 2>/dev/null || true

%_wrap.go: %.i FORCE
	swig -v -c++ -go $<

swig: $(GO_FILES)

.PHONY: test prepare swig build FORCE
