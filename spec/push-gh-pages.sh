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

cp -R ../examples examples ; cp -R ../schema schema
git init
git config user.email "{GH_USER_EMAIL}"
git config user.name "${GH_USER_NAME}"
git add . ; git commit -m "Deploy to GitHub Pages"
git push --force --quiet "https://${GH_TOKEN}@${GH_REF}" master:gh-pages > /dev/null 2>&1