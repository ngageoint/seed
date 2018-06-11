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
    sudo cp ../examples/complete/seed.manifest.json $TRAVIS_TAG/seed.manifest.example.json
    # Make detail.html for old links
    sudo cp $TRAVIS_TAG/index.html $TRAVIS_TAG/detail.html
    sudo cp $TRAVIS_TAG/index.html $TRAVIS_TAG/details.html
fi

# Grab all available versions to place in gh-pages
if [[ "${TRAVIS_TAG}x" == "x" ]]
then
    # Place master contents into master directory
    sudo mkdir -p $OUTPUT_DIR/master/schema
    sudo cp seed.* master/
    sudo cp master/seed.html master/index.html
    sudo cp ../schema/* $OUTPUT_DIR/master/schema

    # Create folder structure
    sudo mkdir -p $OUTPUT_DIR/schema

    # Place versioned spec and schemas
    for VERSION in $(cat ../../.versions)
    do
        cd $OUTPUT_DIR
        sudo mkdir -p $VERSION/schema
        cd $VERSION
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.html
        # We want https://ngageoint.github.io/$VERSION/ to also serve up versioned spec... not require /seed.html
        sudo cp seed.html index.html
        # Make detail.html work too for old links
        sudo cp seed.html detail.html
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.pdf
        cd schema
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.manifest.schema.json
        sudo wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed.metadata.schema.json

        # We are going to place all versions in the root, the last one will win, which satisfies our goal of latest tag
        # being the one seen when hitting the GitHub Pages site.
        cd $OUTPUT_DIR
        sudo cp -rf $VERSION/seed* ./
        sudo cp -rf $VERSION/schema/ schema/
    done
fi

# Chown the entire output directory recursively
sudo chown -R travis:travis $OUTPUT_DIR

popd > /dev/null
