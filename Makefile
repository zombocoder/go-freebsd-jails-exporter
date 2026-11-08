.PHONY: help build test vet fmt fmt-check tidy-check check smoke install-service deps clean dist

GO       ?= go
BINARY   := go-freebsd-jails-exporter
PKG      := github.com/zombocoder/go-freebsd-jails-exporter
PACKAGES := ./...

# Portable between BSD make (FreeBSD's default) and GNU make.
# Overridable, e.g. `make build VERSION=0.3.0 REVISION=abc1234` (CI does this).
_GIT_TAG    != git describe --tags --exact-match 2>/dev/null | sed 's/^v//' || true
_GIT_REV    != git rev-parse --short HEAD 2>/dev/null || echo unknown
_DATE       != date -u +%Y-%m-%dT%H:%M:%SZ
_OS_MAJOR   != freebsd-version -u 2>/dev/null | cut -d. -f1 || true
_ARCH       != uname -m

VERSION    ?= $(_GIT_TAG)
REVISION   ?= $(_GIT_REV)
BUILD_DATE ?= $(_DATE)

LDFLAGS = -s -w \
	-X $(PKG)/pkg/version.Revision=$(REVISION) \
	-X $(PKG)/pkg/version.BuildDate=$(BUILD_DATE)

# Empty VERSION keeps the default from pkg/version.
VERSION_LDFLAG = $${VERSION:+-X $(PKG)/pkg/version.Version=$$VERSION}

help:
	@echo "go-freebsd-jails-exporter"
	@echo "========================="
	@echo ""
	@echo "Available targets:"
	@echo "  build            - Build ./$(BINARY) with version info"
	@echo "  test             - Run unit tests"
	@echo "  vet              - Run go vet"
	@echo "  fmt              - Format code"
	@echo "  fmt-check        - Fail if any file is not gofmt-formatted"
	@echo "  tidy-check       - Fail if go.mod/go.sum are not tidy"
	@echo "  check            - fmt-check, tidy-check, vet and test"
	@echo "  smoke            - End-to-end test with a throwaway pool and jail (root, test machine only)"
	@echo "  install-service  - Install the binary and rc.d script (root)"
	@echo "  deps             - Check build and runtime dependencies"
	@echo "  dist             - Build a release tarball into ./dist"
	@echo "  clean            - Remove build artifacts"

build:
	VERSION="$(VERSION)"; CGO_ENABLED=1 $(GO) build -trimpath -ldflags "$(LDFLAGS) $(VERSION_LDFLAG)" -o $(BINARY) ./cmd/$(BINARY)/

test:
	CGO_ENABLED=1 $(GO) test -count=1 $(PACKAGES)

vet:
	CGO_ENABLED=1 $(GO) vet $(PACKAGES)

fmt:
	gofmt -s -w .

fmt-check:
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then echo "Not gofmt-formatted:"; echo "$$unformatted"; exit 1; fi

tidy-check:
	$(GO) mod tidy -diff

check: fmt-check tidy-check vet test

# Creates and destroys a file-backed pool "jexci" and a jail "jexci".
smoke: build
	@if [ "$$(id -u)" != 0 ]; then echo "smoke must run as root (e.g. doas make smoke)"; exit 1; fi
	sh .github/scripts/freebsd-ci-smoke.sh ./$(BINARY)

install-service: build
	@if [ "$$(id -u)" != 0 ]; then echo "install-service must run as root"; exit 1; fi
	install -m 0755 $(BINARY) /usr/local/bin/$(BINARY)
	install -m 0555 scripts/freebsd/rc.d/go_freebsd_jails_exporter /usr/local/etc/rc.d/go_freebsd_jails_exporter
	@echo "Enable with: sysrc go_freebsd_jails_exporter_enable=YES && service go_freebsd_jails_exporter start"

deps:
	@echo "Checking dependencies..."
	@printf "Go: "; $(GO) version || (echo "not found (pkg install go)"; exit 1)
	@printf "FreeBSD: "; freebsd-version -ku 2>/dev/null | tr '\n' ' ' || (echo "not running on FreeBSD"; exit 1); echo
	@printf "OpenZFS headers: "; [ -f /usr/src/sys/contrib/openzfs/lib/libspl/include/libshare.h ] && echo found || echo "not found (see README: FreeBSD sources)"
	@printf "RACCT: "; [ "$$(sysctl -n kern.racct.enable 2>/dev/null)" = 1 ] && echo enabled || echo "disabled (set kern.racct.enable=1 in /boot/loader.conf and reboot)"

_DIST_VER != v='$(VERSION)'; echo "$${v:-dev}"
DIST_NAME   = $(BINARY)-$(_DIST_VER)-freebsd$(_OS_MAJOR)-$(_ARCH)

dist: build
	rm -rf dist/$(DIST_NAME)
	mkdir -p dist/$(DIST_NAME)
	cp $(BINARY) LICENSE README.md CHANGELOG.md dist/$(DIST_NAME)/
	cp scripts/freebsd/rc.d/go_freebsd_jails_exporter grafana/dashboard.json dist/$(DIST_NAME)/
	tar -czf dist/$(DIST_NAME).tar.gz -C dist $(DIST_NAME)
	cd dist && sha256 -r $(DIST_NAME).tar.gz > $(DIST_NAME).tar.gz.sha256
	@echo "dist/$(DIST_NAME).tar.gz"

clean:
	rm -rf $(BINARY) dist
