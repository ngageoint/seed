#!/usr/bin/env bash

set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

VERSION=$1
if [[ "${VERSION}x" == "x" ]]
then
    echo Missing version parameter!
    echo Usage:
    echo $0 1.0.0
    exit 1
fi


vendor/go-bindata -pkg constants -o constants/jsonschema.go ../spec/schema/seed.manifest.schema.json
echo Building cross platform Seed CLI.
echo Building for Linux...
GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o output/seed-linux-amd64
echo Building for OSX...
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o output/seed-darwin-amd64
echo Building for Windows...
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=$VERSION" -o output/seed-windows-amd64
echo CLI build complete

popd >/dev/null