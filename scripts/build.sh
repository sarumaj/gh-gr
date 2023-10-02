#!/bin/bash
set -e

supported_platforms=(
    darwin-amd64
    darwin-arm64
    freebsd-386
    freebsd-amd64
    freebsd-arm64
    linux-386
    linux-amd64
    linux-arm
    linux-arm64
    windows-386
    windows-amd64
    windows-arm64
)

VERSION="$1"
BUILD_DATE="$(date -u "+%Y-%m-%d %H:%M:%S UTC")"

go install golang.org/x/tools/gopls@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

gofmt -s -d ./
golangci-lint run

for p in "${supported_platforms[@]}"; do
    goos="${p%-*}"
    goarch="${p#*-}"

    ext=""
    if [ "$goos" = "windows" ]; then
        ext=".exe"
    fi

    GOOS="$goos" GOARCH="$goarch" CGO_ENABLED="${CGO_ENABLED:-0}" go build \
        -trimpath \
        -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.BuildDate=${BUILD_DATE}' -extldflags=-static" \
        -tags="osusergo netgo static_build" \
        -o "dist/gh-gr_${VERSION}_${p}${ext}"
done