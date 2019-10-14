#!/usr/bin/env bash

: ${BUILD_TAG:=${TRAVIS_TAG}}
: ${RELEASE_DOWNLOAD_URL:=https://github.com/ngageoint/seed/releases/download/}

# Ensure script directory is CWD
OUTPUT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"/output
pushd ${OUTPUT_DIR}
OUTPUT_DIR=$(pwd)


# Collect files for GitHub Releases
if [[ "${BUILD_TAG}x" != "x" ]]
then
    mkdir ${BUILD_TAG}
    cp -R seed.* ${BUILD_TAG}/
    cp -R ../schema/*.json ${BUILD_TAG}/
    cp ../examples/complete/seed.manifest.json ${BUILD_TAG}/seed.manifest.example.json
fi

# Grab all available versions to place in gh-pages
if [[ "${BUILD_TAG}x" == "x" ]]
then
    # Place master contents into master directory
    mkdir -p ${OUTPUT_DIR}/master/schema
    cp seed.* master/
    cp master/seed.html master/index.html
    cp ../schema/* ${OUTPUT_DIR}/master/schema

    # Create folder structure
    mkdir -p ${OUTPUT_DIR}/schema

    # Place versioned spec and schemas
    for VERSION in $(cat ../../.versions)
    do
        cd ${OUTPUT_DIR}
        mkdir -p ${VERSION}/schema
        cd ${VERSION}
        wget ${RELEASE_DOWNLOAD_URL}${VERSION}/seed.html
        # We want https://ngageoint.github.io/$VERSION/ to also serve up versioned spec... not require /seed.html
        cp seed.html index.html
        wget ${RELEASE_DOWNLOAD_URL}${VERSION}/seed.pdf
        cd schema
        wget ${RELEASE_DOWNLOAD_URL}${VERSION}/seed.manifest.schema.json
        wget ${RELEASE_DOWNLOAD_URL}${VERSION}/seed.metadata.schema.json

        # We are going to place all versions in the root, the last one will win, which satisfies our goal of latest tag
        # being the one seen when hitting the GitHub Pages site.
        cd ${OUTPUT_DIR}
	    # Place seed.html at detail.html for purposes of legacy shared links
        cp -rf ${VERSION}/seed.html ./detail.html
        cp -rf ${VERSION}/seed* ./
        cp -rf ${VERSION}/schema/ ./
    done
fi

# Chown the entire output directory recursively
chmod -R 755 ${OUTPUT_DIR}

popd > /dev/null
