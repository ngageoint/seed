#!/usr/bin/env bash

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

vendor/bin/go-bindata -pkg constants -o constants/jsonschema.go ../spec/schema/seed.manifest.schema.json
echo Building cross platform Seed CLI.
echo Building for Linux...
GOOS=linux GOARCH=amd64 go build -o output/seed-linux-amd64
echo Building for OSX...
GOOS=darwin GOARCH=amd64 go build -o output/seed-darwin-amd64
echo Building for Windows...
GOOS=windows GOARCH=amd64 go build -o output/seed-windows-amd64
echo CLI build complete

popd >/dev/null