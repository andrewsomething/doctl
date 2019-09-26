# NOTE: Adding a target so it shows up in the help listing
#    - The description is the text that is echoed in the first command in the target.
#    - Only 'public' targets (start with an alphanumeric character) display in the help listing.
#    - All public targets need a description

export CGO = 0

export GO111MODULE := on

# ORIGIN is used when testing release code
ORIGIN ?= origin
BUMP ?= patch

.PHONY: help
help:
	@echo "==> describe make commands"
	@echo ""
	@$(MAKE) -pRrq -f $(lastword $(MAKEFILE_LIST)) : 2>/dev/null |\
	  awk -v RS= -F: \
	    '/^# File/,/^# Finished Make data base/ {if ($$1 ~ /^[a-zA-Z]/) {printf "%-20s%s\n", $$1, substr($$9, 9, length($$9)-9)}}' |\
	  sort

my_d = $(shell pwd)
OUT_D = $(shell echo $${OUT_D:-$(my_d)/builds})

UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -p)

GOOS = linux
ifeq ($(UNAME_S),Darwin)
  GOOS = darwin
endif

GOARCH = amd64
ifneq ($(UNAME_P), x86_64)
  GOARCH = 386
endif

.PHONY: _build
_build:
	@echo "=> building doctl via go build"
	@echo ""
	@mkdir -p builds
	@cd cmd/doctl && env GOOS=$(GOOS) GOARCH=$(GOARCH) GOFLAGS=-mod=vendor \
	  go build -o $(OUT_D)/doctl_$(GOOS)_$(GOARCH)
	@echo "built $(OUT_D)/doctl_$(GOOS)_$(GOARCH)"

.PHONY: native
native: _build
	@echo "==> build local version"
	@echo ""
	@mv $(OUT_D)/doctl_$(GOOS)_$(GOARCH) $(OUT_D)/doctl
	@echo "built $(OUT_D)/doctl"

.PHONY: _build_linux_amd64
_build_linux_amd64: GOOS = linux
_build_linux_amd64: GOARCH = amd64
_build_linux_amd64: _build

.PHONY: _base_docker_cntr
_base_docker_cntr:
	@docker build -f Dockerfile.build . -t doctl_builder

.PHONY: docker_build
docker_build: _base_docker_cntr
	@echo "==> build doctl in local docker container"
	@echo ""
	@mkdir -p $(OUT_D)
	@docker build -f Dockerfile.cntr . -t doctl_local
	@docker run --rm \
		-v $(OUT_D):/copy \
		-it --entrypoint /usr/bin/rsync \
		doctl_local -av /app/ /copy/
	@docker run --rm \
		-v $(OUT_D):/copy \
		-it --entrypoint /bin/chown \
		alpine -R $(shell whoami | id -u): /copy
	@echo "Built binaries to $(OUT_D)"
	@echo "Created a local Docker container. To use, run: docker run --rm -it doctl_local"

.PHONY: test_unit
test_unit:
	@echo "==> running only unit tests"
	@echo ""
	go test ./commands/... ./do/... ./pkg/... .

.PHONY: test_integration
test_integration:
	@echo "==> running just integration tests"
	@echo ""
	go test ./integration

.PHONY: test
test: test_unit test_integration

.PHONY: shellcheck
shellcheck:
	@echo "==> analyze shell scripts"
	@echo ""
	@scripts/shell_check.sh

CHANNEL ?= stable

.PHONY: _snap
_snap:
	@echo "=> publishing snap"
	@echo ""
	@CHANNEL=${CHANNEL} scripts/snap.sh

.PHONY: mocks
mocks:
	@echo "==> update mocks"
	@echo ""
	@scripts/regenmocks.sh

.PHONY: vendor
vendor:
	@echo "==> vendor dependencies"
	@echo ""
	go mod vendor
	go mod tidy

.PHONY: clean
clean:
	@echo "==> remove build / release artifacts"
	@echo ""
	@rm -rf builds dist out

.PHONY: _install_github_release_notes
_install_github_release_notes:
	@GO111MODULE=off go get -u github.com/buchanae/github-release-notes

.PHONY: _changelog
_changelog: _install_github_release_notes
	@scripts/changelog.sh

.PHONY: changes
changes: _install_github_release_notes
	@echo "==> list merged PRs since last release"
	@echo ""
	@changes=$(shell scripts/changelog.sh) && cat $$changes && rm -f $$changes

.PHONY: version
version:
	@echo "==> doctl version"
	@echo ""
	@ORIGIN=${ORIGIN} scripts/version.sh

.PHONY: _install_sembump
_install_sembump:
	@echo "=> installing/updating sembump tool"
	@echo ""
	@GO111MODULE=off go get -u github.com/jessfraz/junk/sembump

.PHONY: _bump_and_tag
_bump_and_tag: _install_sembump
	@echo "=> BUMP=${BUMP} bumping and tagging version"
	@echo ""
	@ORIGIN=${ORIGIN} scripts/bumpversion.sh

.PHONY: _release
_release:
	@echo "=> releasing"
	@echo ""
	@scripts/release.sh

.PHONY: bump_and_release
bump_and_release: _bump_and_tag
	@echo "==> BUMP=${BUMP} bump tag and release"
	@echo ""
	@$(MAKE) _release
	@$(MAKE) _snap

.PHONY: release
release:
	@echo "==> release most recent tag"
	@echo ""
	@$(MAKE) _release


