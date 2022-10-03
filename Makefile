VERSION = $(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
OSARCH=$(shell go env GOHOSTOS)-$(shell go env GOHOSTARCH)

DEPTOKENS=\
	deptokens-darwin-arm64 \
	deptokens-darwin-amd64 \
	deptokens-linux-amd64 \
	deptokens-windows-amd64.exe

DEPSERVER=\
	depserver-darwin-arm64 \
	depserver-darwin-amd64 \
	depserver-linux-amd64 \
	depserver-windows-amd64.exe

DEPSYNCER=\
	depsyncer-darwin-arm64 \
	depsyncer-darwin-amd64 \
	depsyncer-linux-amd64 \
	depsyncer-windows-amd64.exe

SUPPLEMENTAL=\
	tools/*.sh \
	docs/dep-profile.example.json

my: deptokens-$(OSARCH) depserver-$(OSARCH) depsyncer-$(OSARCH)

docker: depserver-linux-amd64 depsyncer-linux-amd64

$(DEPTOKENS): cmd/deptokens
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

$(DEPSERVER): cmd/depserver
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

$(DEPSYNCER): cmd/depsyncer
	GOOS=$(word 2,$(subst -, ,$@)) GOARCH=$(word 3,$(subst -, ,$(subst .exe,,$@))) go build $(LDFLAGS) -o $@ ./$<

nanodep-%-$(VERSION).zip: depserver-% depsyncer-% deptokens-% $(SUPPLEMENTAL)
	rm -rf $@ $(subst .zip,,$@)
	mkdir $(subst .zip,,$@)
	echo $^ | xargs -n 1 | cpio -pdmu $(subst .zip,,$@)
	zip -r $@ $(subst .zip,,$@)
	rm -rf $(subst .zip,,$@)

nanodep-%-$(VERSION).zip: depserver-%.exe depsyncer-%.exe deptokens-%.exe $(SUPPLEMENTAL)
	rm -rf $@ $(subst .zip,,$@)
	mkdir $(subst .zip,,$@)
	echo $^ | xargs -n 1 | cpio -pdmu $(subst .zip,,$@)
	zip -r $@ $(subst .zip,,$@)
	rm -rf $(subst .zip,,$@)

clean:
	rm -f deptokens-* depserver-* depsyncer-* nanodep-*.zip

release: \
	nanodep-darwin-amd64-$(VERSION).zip \
	nanodep-darwin-arm64-$(VERSION).zip \
	nanodep-linux-amd64-$(VERSION).zip

test:
	go test -v -cover -race ./...

.PHONY: my docker $(DEPTOKENS) $(DEPSERVER) $(DEPSYNCER) clean release test