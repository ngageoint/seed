#!/usr/bin/env bash

# Ensure script directory is CWD
cd "${0%/*}"/output

if [[ "${TRAVIS_TAG}x" != "x" ]]
then
    cp detail.html seed-${TRAVIS_TAG}.html
    cp detail.pdf seed-${TRAVIS_TAG}.pdf;
fi

# Grab all available versions to place in gh-pages
for VERSION in $(cat ../../.versions)
do
    wget https://github.com/ngageoint/seed/releases/download/${VERSION}/seed-${VERSION}.html
done

cd - > /dev/null