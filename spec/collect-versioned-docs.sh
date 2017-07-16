#!/usr/bin/env bash

# Ensure script directory is CWD
OUTPUT_DIR="${0%/*}"/output
pushd $OUTPUT_DIR

if [[ "${TRAVIS_TAG}x" != "x" ]]
then
    sudo mkdir ${TRAVIS_TAG}
    sudo cp -R seed.* $TRAVIS_TAG/
    sudo cp -R schema/*.json $TRAVIS_TAG/
fi

# Grab all available versions to place in gh-pages
for VERSION in $(cat ../../.versions)
do
    cd $OUTPUT_DIR
    sudo mkdir -p $VERSION/schema
    cd $VERSION
    sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.html
    sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.pdf
    cd $VERSION/schema
    sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.manifest.schema.json
    sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.metadata.schema.json
done

popd > /dev/null