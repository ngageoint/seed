#!/usr/bin/env bash

# Ensure script directory is CWD
cd "${0%/*}"

# Only push docs to gh-pages on explicit builds from master. Pull requests will be ignored
if [ "${TRAVIS_BRANCH}" != "master" ] || [ "${TRAVIS_PULL_REQUEST}" != "false" ]; then
  echo "Skipping update of Github hosted documentation. Updates are made only on builds of master branch - pull requests are skipped."
  exit 0
fi

cd output ; cp -R ../examples examples ; cp -R ../schema schema

git init
git config user.name "${GH_USER_NAME}"
git config user.email "{GH_USER_EMAIL}"
git add . ; git commit -m "Deploy to GitHub Pages"
git push --force --quiet "https://${GH_TOKEN}@${GH_REF}" master:gh-pages > /dev/null 2>&1