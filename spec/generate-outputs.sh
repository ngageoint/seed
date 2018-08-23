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

# Compile css from sass
docker run --rm -v $(pwd):/var/www ${SASS_IMAGE} sh styles/compile-sass.sh

# Generate HTML
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -D /documents/output seed.adoc

# Patch doc for PDF generation
perl -0777 -i.bak -pe 's/\/\/{pdf\+4}\na/4\+a/g;' -pe 's/\/\/{pdf\+1}\n3\+a/4\+a/g' $(pwd)/sections/standard.adoc

# Generate PDF
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor-pdf -a pdf-style=styles/pdf-theme.yml -a pdf-fontsdir=styles/fonts/ -D /documents/output seed.adoc

# Generate manpage styled adoc
docker run -v $(pwd):/documents --rm ${PYTHON_IMAGE} python /documents/generate-manpage.py

# Generate manpage
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -b manpage -D /documents/output seed.man.adoc

# Replace original doc following PDF generation
mv $(pwd)/sections/standard.adoc.bak $(pwd)/sections/standard.adoc

popd > /dev/null