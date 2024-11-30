SHELL :=/bin/bash -e -o pipefail
PWD   := $(shell pwd)

BUILD_MODE?=c-shared
OUTPUT_DIR?=output
GO_BINARY?=go
BINDING_NAME?=slashengine
BINDING_FILE?=$(BINDING_NAME).so
BINDING_ARGS?=
BINDING_OUTPUT?=$(OUTPUT_DIR)/binding
EXTRA_LD_FLAGS?=

.DEFAULT_GOAL := all
.PHONY: all
all: ## build pipeline
all: mod inst gen build spell lint test

.PHONY: precommit
precommit: ## validate the branch before commit
precommit: all vuln

.PHONY: ci
ci: ## CI build pipeline
ci: lint-reports test vuln precommit diff

.PHONY: help
help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: out
out: ## create out directory
	@mkdir -p out

.PHONY: git-hooks
git-hooks: ## install git hooks
	@git config --local core.hooksPath .githooks/

.PHONY: deps
deps: ## Downloads the dependencies
	@go mod download

.PHONY: run
run: fmt ## Run the app
	@go run ./cmd/main.go

.PHONY: test-build
test-build: ## Tests whether the code compiles
	@go build -o /dev/null ./...

.PHONY: clean
clean: ## remove files created during build pipeline
	$(call print-target)
	@rm -rf dist bin out build output
	@rm -f coverage.*
	@rm -f '"$(shell go env GOCACHE)/../golangci-lint"'
	@go clean -i -cache -testcache -modcache -fuzzcache -x

.PHONY: mod
mod: ## go mod tidy, cleans up go.mod and go.sum
	$(call print-target)
	@go mod tidy
	@cd tools && go mod tidy

.PHONY: fmt
fmt: ## Formats all code with go fmt
	@go fmt ./...

.PHONY: inst
inst: ## go install tools
	$(call print-target)
	@cd tools && go install $(shell cd tools && go list -e -f '{{ join .Imports " " }}' -tags=tools)

.PHONY: get
get: ## get and update dependencies
	$(call print-target)
	@go get -u ./...

.PHONY: gen
gen: ## go generate
	$(call print-target)
	@go generate ./...

.PHONY: build
build: ## goreleaser build
	$(call print-target)
	@goreleaser build --clean --single-target --snapshot

.PHONY: spell
spell: ## misspell
	$(call print-target)
	@misspell -error -locale=US -w **.md

.PHONY: lint
lint: fmt deps ## Lints all code with golangci-lint
	$(call print-target)
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --fix

.PHONY: lint-reports
lint-reports: out deps ## Lint reports
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run ./... --out-format checkstyle | tee "$(@)"

.PHONY: vuln
vuln: ## govulncheck
	$(call print-target)
	@go install golang.org/x/vuln/cmd/govulncheck@latest
	@govulncheck ./...

.PHONY: test
test: ## go test
	$(call print-target)
	@go test -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./... ./...
	@go tool cover -html=coverage.out -o coverage.html

.PHONY: diff
diff: ## git diff
	$(call print-target)
	@git diff --exit-code
	@RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

.PHONY: binding
binding: deps ## build the binding
	mkdir -p $(BINDING_OUTPUT)
	$(GO_BINARY) build -ldflags="-w -s $(EXTRA_LD_FLAGS)" -o $(BINDING_OUTPUT)/$(BINDING_FILE) -buildmode=$(BUILD_MODE) $(BINDING_ARGS) binding/main.go

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef

# Android

ANDROID_HOME?=$(HOME)/Android/Sdk
ANDROID_NDK_HOME?=$(ANDROID_HOME)/ndk/21.3.6528147
ANDROID_NDK_TOOLCHAIN?=$(ANDROID_NDK_HOME)/toolchains/llvm/prebuilt/linux-x86_64/bin
ANDROID_OUTPUT?=android/jniLibs
ANDROID_BINDING_NAME?=$(BINDING_NAME).so

.PHONY: binding_android
binding_android: binding_android_arm64 binding_android_armv7a binding_android_x86 binding_android_x86_64

.PHONY: binding_android_arm64
binding_android_arm64:
	BINDING_FILE=$(ANDROID_OUTPUT)/arm64-v8a/$(ANDROID_BINDING_NAME) \
	CC=$(ANDROID_NDK_TOOLCHAIN)/aarch64-linux-android21-clang \
	EXTRA_LD_FLAGS="-extldflags=-Wl,-soname,$(ANDROID_BINDING_NAME)" \
	GOOS=android GOARCH=arm64 CGO_ENABLED=1 make binding

.PHONY: binding_android_armv7a
binding_android_armv7a:
	BINDING_FILE=$(ANDROID_OUTPUT)/armeabi-v7a/$(ANDROID_BINDING_NAME) \
	CC=$(ANDROID_NDK_TOOLCHAIN)/armv7a-linux-androideabi21-clang \
	EXTRA_LD_FLAGS="-extldflags=-Wl,-soname,$(ANDROID_BINDING_NAME)" \
	GOOS=android GOARCH=arm GOARM=7 CGO_ENABLED=1 make binding

.PHONY: binding_android_x86
binding_android_x86:
	BINDING_FILE=$(ANDROID_OUTPUT)/x86/$(ANDROID_BINDING_NAME) \
	CC=$(ANDROID_NDK_TOOLCHAIN)/i686-linux-android21-clang \
	EXTRA_LD_FLAGS="-extldflags=-Wl,-soname,$(ANDROID_BINDING_NAME)" \
	GOOS=android GOARCH=386 CGO_ENABLED=1 make binding

.PHONY: binding_android_x86_64
binding_android_x86_64:
	BINDING_FILE=$(ANDROID_OUTPUT)/x86_64/$(ANDROID_BINDING_NAME) \
	CC=$(ANDROID_NDK_TOOLCHAIN)/x86_64-linux-android21-clang \
	EXTRA_LD_FLAGS="-extldflags=-Wl,-soname,$(ANDROID_BINDING_NAME)" \
	GOOS=android GOARCH=amd64 CGO_ENABLED=1 make binding

# macOS / Darwin

DARWIN_OUTPUT?=darwin
DARWIN_BINDING_OUTPUT?=$(BINDING_OUTPUT)/$(DARWIN_OUTPUT)
DARWIN_TARGET?=10.11
DARWIN_SDKROOT?=$(shell xcrun --sdk macosx --show-sdk-path)

.PHONY: binding_darwin
binding_darwin: binding_darwin_x86_64 binding_darwin_arm64
	lipo $(DARWIN_BINDING_OUTPUT)/x86_64/$(BINDING_NAME).dylib $(DARWIN_BINDING_OUTPUT)/arm64/$(BINDING_NAME).dylib -create -output $(DARWIN_BINDING_OUTPUT)/$(BINDING_NAME).dylib
	rm -rf $(DARWIN_BINDING_OUTPUT)/x86_64/$(BINDING_NAME).dylib $(DARWIN_BINDING_OUTPUT)/arm64/$(BINDING_NAME).dylib $(DARWIN_BINDING_OUTPUT)/arm64 $(DARWIN_BINDING_OUTPUT)/x86_64

.PHONY: binding_darwin_x86_64
binding_darwin_x86_64:
	BINDING_FILE=$(DARWIN_OUTPUT)/x86_64/$(BINDING_NAME).dylib \
	BUILD_MODE="c-shared" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	make binding
	install_name_tool -id @rpath/$(BINDING_NAME).dylib $(DARWIN_BINDING_OUTPUT)/x86_64/$(BINDING_NAME).dylib

.PHONY: binding_darwin_arm64
binding_darwin_arm64:
	BINDING_FILE=$(DARWIN_OUTPUT)/arm64/$(BINDING_NAME).dylib \
	BUILD_MODE="c-shared" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	SDKROOT=$(DARWIN_SDKROOT) \
	make binding
	install_name_tool -id @rpath/$(BINDING_NAME).dylib $(DARWIN_BINDING_OUTPUT)/arm64/$(BINDING_NAME).dylib

.PHONY: binding_darwin_archive_x86_64
binding_darwin_archive_x86_64:
	BINDING_FILE=$(DARWIN_OUTPUT)/x86_64.a \
	BUILD_MODE="c-archive" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_darwin_archive_arm64
binding_darwin_archive_arm64:
	BINDING_FILE=$(DARWIN_OUTPUT)/arm64.a \
	BUILD_MODE="c-archive" \
	CGO_CFLAGS=-mmacosx-version-min=$(DARWIN_TARGET) \
	MACOSX_DEPLOYMENT_TARGET=$(DARWIN_TARGET) \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	make binding

# iOS

IOS_OUTPUT?=ios
IOS_BINDING_OUTPUT?=$(BINDING_OUTPUT)/$(IOS_OUTPUT)
IOS_BINDING_NAME?=$(BINDING_NAME).a

.PHONY: binding_ios
binding_ios: binding_ios_arm64 binding_ios_x86_64_sim
	lipo $(IOS_BINDING_OUTPUT)/x86_64_sim/$(IOS_BINDING_NAME) $(IOS_BINDING_OUTPUT)/arm64/$(IOS_BINDING_NAME) $(IOS_BINDING_OUTPUT)/armv7/$(IOS_BINDING_NAME) -create -output $(IOS_BINDING_OUTPUT)/$(IOS_BINDING_NAME)
	cp $(IOS_BINDING_OUTPUT)/arm64/*.h $(IOS_BINDING_OUTPUT)
	rm -rf $(IOS_BINDING_OUTPUT)/arm64 $(IOS_BINDING_OUTPUT)/x86_64_sim $(IOS_BINDING_OUTPUT)/armv7

.PHONY: binding_ios_xcframework
binding_ios_xcframework: binding_ios_all_iphone binding_ios_all_sim binding_ios_all_catalyst
	mkdir -p $(IOS_BINDING_OUTPUT)/headers
	cp $(IOS_BINDING_OUTPUT)/arm64/*.h $(IOS_BINDING_OUTPUT)/headers
	rm -rf $(IOS_BINDING_OUTPUT)/Rsa.xcframework
	xcodebuild -create-xcframework \
		-library $(IOS_BINDING_OUTPUT)/arm64/$(IOS_BINDING_NAME) -headers $(IOS_BINDING_OUTPUT)/headers \
		-library $(IOS_BINDING_OUTPUT)/sim/$(IOS_BINDING_NAME) -headers $(IOS_BINDING_OUTPUT)/headers \
		-library $(IOS_BINDING_OUTPUT)/catalyst/$(IOS_BINDING_NAME) -headers $(IOS_BINDING_OUTPUT)/headers \
		-output $(IOS_BINDING_OUTPUT)/Rsa.xcframework
	rm -rf $(IOS_BINDING_OUTPUT)/arm64 $(IOS_BINDING_OUTPUT)/sim $(IOS_BINDING_OUTPUT)/catalyst $(IOS_BINDING_OUTPUT)/headers

.PHONY: binding_ios_all_iphone
binding_ios_all_iphone: binding_ios_arm64

.PHONY: binding_ios_all_sim
binding_ios_all_sim: binding_ios_x86_64_sim binding_ios_arm64_sim
	mkdir -p $(IOS_BINDING_OUTPUT)/sim
	lipo $(IOS_BINDING_OUTPUT)/x86_64_sim/$(IOS_BINDING_NAME) $(IOS_BINDING_OUTPUT)/arm64_sim/$(IOS_BINDING_NAME) -create -output $(IOS_BINDING_OUTPUT)/sim/$(IOS_BINDING_NAME)
	rm -rf $(IOS_BINDING_OUTPUT)/x86_64_sim $(IOS_BINDING_OUTPUT)/arm64_sim

.PHONY: binding_ios_all_catalyst
binding_ios_all_catalyst: binding_ios_x86_64_catalyst binding_ios_arm64_catalyst
	mkdir -p $(IOS_BINDING_OUTPUT)/catalyst
	lipo $(IOS_BINDING_OUTPUT)/x86_64_catalyst/$(IOS_BINDING_NAME) $(IOS_BINDING_OUTPUT)/arm64_catalyst/$(IOS_BINDING_NAME) -create -output $(IOS_BINDING_OUTPUT)/catalyst/$(IOS_BINDING_NAME)
	rm -rf $(IOS_BINDING_OUTPUT)/x86_64_catalyst $(IOS_BINDING_OUTPUT)/arm64_catalyst

.PHONY: binding_ios_x86_64_catalyst
binding_ios_x86_64_catalyst:
	CGO_LDFLAGS="-target x86_64-apple-ios14-macabi" \
	BINDING_FILE=$(IOS_OUTPUT)/x86_64_catalyst/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=macosx CC=$(PWD)/clangwrap.sh \
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_ios_arm64_catalyst
binding_ios_arm64_catalyst:
	CGO_LDFLAGS="-target arm64-apple-ios14-macabi -fembed-bitcode" \
	BINDING_FILE=$(IOS_OUTPUT)/arm64_catalyst/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=macosx CC=$(PWD)/clangwrap.sh \
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_ios_x86_64_sim
binding_ios_x86_64_sim:
	CGO_LDFLAGS="-target x86_64-apple-ios10-simulator -fembed-bitcode" \
	BINDING_FILE=$(IOS_OUTPUT)/x86_64_sim/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=iphonesimulator CC=$(PWD)/clangwrap.sh \
	GOOS=ios GOARCH=amd64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_ios_arm64_sim
binding_ios_arm64_sim:
	CGO_LDFLAGS="-target arm64-apple-ios10-simulator -fembed-bitcode" \
	BINDING_FILE=$(IOS_OUTPUT)/arm64_sim/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=iphonesimulator CC=$(PWD)/clangwrap.sh \
	GOOS=ios GOARCH=arm64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_ios_arm64
binding_ios_arm64:
	CGO_LDFLAGS="-target arm64-apple-ios10 -fembed-bitcode" \
	BINDING_FILE=$(IOS_OUTPUT)/arm64/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=iphoneos CC=$(PWD)/clangwrap.sh \
	GOOS=ios GOARCH=arm64 CGO_ENABLED=1 \
	make binding

.PHONY: binding_ios_armv7
binding_ios_armv7:
	BINDING_FILE=$(IOS_OUTPUT)/armv7/$(IOS_BINDING_NAME) BUILD_MODE="c-archive" \
	SDK=iphoneos CC=$(PWD)/clangwrap.sh CGO_CFLAGS="-fembed-bitcode" \
	GOOS=darwin GOARCH=arm CGO_ENABLED=1 BINDING_ARGS="-tags ios" \
	make binding

# Linux

LINUX_OUTPUT?=linux
LINUX_BINDING_NAME?=$(BINDING_NAME).so

.PHONY: binding_linux
binding_linux: binding_linux_386 binding_linux_amd64 binding_linux_arm64

.PHONY: binding_linux_386
binding_linux_386:
	GOOS=linux GOARCH=386 TAG=main \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/386/$(LINUX_BINDING_NAME)" \
	CMD="make binding" ./build.sh

.PHONY: binding_linux_amd64
binding_linux_amd64:
	GOOS=linux GOARCH=amd64 TAG=main \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/amd64/$(LINUX_BINDING_NAME)" \
	CMD="make binding" ./build.sh

.PHONY: binding_linux_arm64
binding_linux_arm64:
	GOOS=linux GOARCH=arm64 TAG=arm \
	ARGS="-e BINDING_FILE=$(LINUX_OUTPUT)/arm64/$(LINUX_BINDING_NAME)" \
	CMD="make binding" ./build.sh

# Windows

WINDOWS_OUTPUT?=windows
WINDOWS_BINDING_NAME?=$(BINDING_NAME).dll

.PHONY: binding_windows
binding_windows: binding_windows_386 binding_windows_amd64

.PHONY: binding_windows_386
binding_windows_386:
	GOOS=windows GOARCH=386 \
	ARGS="-e BINDING_FILE=$(WINDOWS_OUTPUT)/386/$(WINDOWS_BINDING_NAME)" \
	TAG=main CMD="make binding" ./build.sh

.PHONY: binding_windows_amd64
binding_windows_amd64:
	GOOS=windows GOARCH=amd64 TAG=main \
	ARGS="-e BINDING_FILE=$(WINDOWS_OUTPUT)/amd64/$(WINDOWS_BINDING_NAME)" \
	CMD="make binding" ./build.sh

# WebAssembly

TINYGO_ROOT?=`tinygo env TINYGOROOT`
GO_ROOT?=`go env GOROOT`

.PHONY: wasm_tinygo
wasm_tinygo:
	mkdir -p output/wasm
	tinygo build -tags=math_big_pure_go -o output/wasm/rsa.wasm -target wasm wasm/main.go
	cp $(TINYGO_ROOT)/targets/wasm_exec.js  output/wasm/wasm_exec.js
	cp output/wasm/rsa.wasm wasm/example/public/rsa.wasm
	cp output/wasm/wasm_exec.js  wasm/example/public/wasm_exec.js

.PHONY: wasm
wasm:
	mkdir -p output/wasm
	cd wasm && GOARCH=wasm GOOS=js go build -ldflags="-s -w" -o ../output/wasm/rsa.wasm main.go
	cp $(GO_ROOT)/misc/wasm/wasm_exec.js  output/wasm/wasm_exec.js
	cp output/wasm/rsa.wasm wasm/example/public/rsa.wasm
	cp output/wasm/wasm_exec.js  wasm/example/public/wasm_exec.js

# GoMobile

GOMOBILE_BRIDGE_PACKAGE?=github.com/jerson/rsa-mobile/bridge
GOMOBILE_BRIDGE_NAME?=RsaBridge
GOMOBILE_PACKAGE?=github.com/jerson/rsa-mobile/rsa
GOMOBILE_NAME?=Rsa

.PHONY: gomobile
gomobile:
	go install golang.org/x/mobile/cmd/gomobile@latest
	gomobile init

.PHONY: gomobile_bridge_android
gomobile_bridge_android:
	mkdir -p output/android
	gomobile bind -ldflags="-w -s" -target=android -o output/android/$(GOMOBILE_BRIDGE_NAME).aar $(GOMOBILE_BRIDGE_PACKAGE)

.PHONY: gomobile_bridge_ios
gomobile_bridge_ios:
	mkdir -p output/ios
	gomobile bind -ldflags="-w -s" -target=ios,iossimulator,macos,maccatalyst -iosversion=14 -o output/ios/$(GOMOBILE_BRIDGE_NAME).xcframework $(GOMOBILE_BRIDGE_PACKAGE)

.PHONY: gomobile_android
gomobile_android:
	mkdir -p output/android
	gomobile bind -ldflags="-w -s" -target=android -o output/android/$(GOMOBILE_NAME).aar $(GOMOBILE_PACKAGE)

.PHONY: gomobile_ios
gomobile_ios:
	mkdir -p output/ios
	gomobile bind -ldflags="-w -s" -target=ios,iossimulator,macos,maccatalyst -iosversion=14 -o output/ios/$(GOMOBILE_NAME).xcframework $(GOMOBILE_PACKAGE)
