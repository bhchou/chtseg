ifeq ($(PHPCFG),)
	PHPCFG=/usr/bin/php-config
endif

ifeq ($(APP),)
	APP=examples
endif

PHPEXE := $(shell $(PHPCFG) --php-binary)
PHPDIR := $(shell $(PHPCFG) --prefix)
PHPEXTDIR := $(shell $(PHPCFG) --extension-dir)

export PATH := $(PHPDIR)/bin:$(PATH)
export CFLAGS := $(shell $(PHPCFG) --includes)
export LDFLAGS := -L$(shell $(PHPCFG) --prefix)/lib/

export CGO_CFLAGS := $(CFLAGS) $(CGO_CFLAGS)
export CGO_LDFLAGS := $(LDFLAGS) $(CGO_LDFLAGS)

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

chtseg:
	go build -v -buildmode=c-shared -o php-chtseg.so php-chtseg.go

chtseg-install:
	cp -f php-chtseg.so $(PHPEXTDIR)

chtseg-test:
	$(PHPEXE) -d extension=./php-chtseg.so ./test.php

chtseg-clean:
	rm -f php-chtseg.so
