#!/usr/bin/env bash

# Description: Creates all supported output document formats
# Usage: ./generate-outputs.sh
# Requires: Docker and perl installed locally
# Variables:
# ASCIDOCTOR_IMAGE: optional override for Asciidoctor Docker image
# PYTHON_IMAGE: optional override for Python 2.7.x Docker image

: ${ASCIIDOCTOR_IMAGE:=rochdev/alpine-asciidoctor:mini}
: ${PYTHON_IMAGE:=python:2-alpine}

# Generate HTML
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -D /documents/output seed.adoc

# Patch doc for PDF generation
perl -0777 -i.bak -pe 's/\/\/{pdf\+4}\na/4\+a/g;' -pe 's/\/\/{pdf\+1}\n3\+a/4\+a/g' $(pwd)/sections/standard.adoc

# Generate PDF
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor-pdf -a pdf-style=style/pdf-theme.yml -a pdf-fontsdir=style/ -D /documents/output seed.adoc

# Generate manpage styled adoc
docker run -v $(pwd):/documents --rm ${PYTHON_IMAGE} python /documents/generate-manpage.py

# Generate manpage
docker run -v $(pwd):/documents --rm ${ASCIIDOCTOR_IMAGE} asciidoctor -b manpage -D /documents/output seed.adoc

# Replace original doc following PDF generation
mv $(pwd)/sections/standard.adoc.bak $(pwd)/sections/standard.adoc