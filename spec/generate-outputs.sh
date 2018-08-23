#!/usr/bin/env bash

# Description: Creates all supported output document formats
# Usage: ./generate-outputs.sh
# Requires: Docker and perl installed locally
# Variables:
# ALPINE_IMAGE: optional override for Alpine Docker image
# ASCIDOCTOR_IMAGE: optional override for Asciidoctor Docker image
# PYTHON_IMAGE: optional override for Python 2.7.x Docker image
# SASS_IMAGE: optional override for Node SASS Docker image

: ${ALPINE_IMAGE:=alpine}
: ${ASCIIDOCTOR_IMAGE:=rochdev/alpine-asciidoctor:mini}
: ${PYTHON_IMAGE:=python:2-alpine}
: ${SASS_IMAGE:=catchdigital/node-sass}

pushd $(dirname $0) > /dev/null

echo Compiling css from sass...
docker run --rm -v $(pwd):/var/www ${SASS_IMAGE} sh styles/compile-sass.sh

echo Generating HTML...
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -D /documents/output seed.adoc

echo Generating PDF...
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} sh generate-pdf.sh

echo Generating manpage styled adoc...
docker run -v $(pwd):/documents --rm ${PYTHON_IMAGE} python /documents/generate-manpage.py

echo Generating manpage...
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -b manpage -D /documents/output seed.man.adoc

popd > /dev/null