#!/usr/bin/env bash

# Ensure script directory is CWD
OUTPUT_DIR="${0%/*}"/output
pushd $OUTPUT_DIR
OUTPUT_DIR=`pwd`


# Collect files for GitHub Releases
if [[ "${TRAVIS_TAG}x" != "x" ]]
then
    sudo mkdir ${TRAVIS_TAG}
    sudo cp -R seed.* $TRAVIS_TAG/
    sudo cp -R ../schema/*.json $TRAVIS_TAG/
    sudo cp ../../cli/output/* $TRAVIS_TAG/
fi

# Grab all available versions to place in gh-pages
if [[ "${TRAVIS_TAG}x" == "x" ]]
then
    # Place snapshot schemas
    sudo mkdir -p $OUTPUT_DIR/schema
    sudo cp ../schema/* $OUTPUT_DIR/schema

    # Place versioned spec and schemas
    for VERSION in $(cat ../../.versions)
    do
        cd $OUTPUT_DIR
        sudo mkdir -p $VERSION/schema
        cd $VERSION
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.html
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.pdf
        cd schema
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.manifest.schema.json
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.metadata.schema.json
    done
fi

# Chown the entire output directory recursively
sudo chown -R travis:travis $OUTPUT_DIR

popd > /dev/null