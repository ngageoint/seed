#!/usr/bin/env bash

## Usage: ./generate-release.sh 1.0.0 [default-branch]

VERSION=$1

if [[ "${VERSION}x" == "x" ]]
then
    echo Missing version parameter!
    echo Usage:
    echo   ./generate-release.sh 1.0.0
    exit 1
fi

DEFAULT_BRANCH=master
if [[ "${2}x" != "x" ]]
then
    DEFAULT_BRANCH=$2
    echo Updated default branch to $2
fi

if [[ "$(which -s gren; echo $?)" == 1 ]]
then
    echo Missing gren application!
    echo Changelog generation requires gren.
    echo Install with: npm install -g gren
fi

# Ensure script directory is CWD
cd "${0%/*}"

tput setaf 2
echo "Building release $VERSION"
tput sgr0

if [[ $(git rev-parse --abbrev-ref HEAD) != "${DEFAULT_BRANCH}" ]]; then
    tput setaf 1
    echo "Current branch is not ${DEFAULT_BRANCH}!"
    tput sgr0
    git rev-parse --abbrev-ref HEAD
    exit 1
fi

git diff-index --quiet HEAD
if [[ $? != 0 ]]; then
    tput setaf 1
    echo "Current index is not clean!"
    tput sgr0
    git diff-index HEAD
fi

tput setaf 2
echo -e "\nDetach the head"
tput sgr0
git checkout --detach

tput setaf 2
echo -e "\nChange the revision on the release, inject manifests, add version history, and create changelog"
tput sgr0
./set-version.sh $VERSION
./spec/examples/inject-manifests.sh
echo $VERSION >> .versions
gren changelog -G -f temp.md -M $VERSION
cat temp.md <(echo) <(cat CHANGELOG.md | sed 's^# Changelog^---^') > temp2.md
rm temp.md
mv temp2.md CHANGELOG.md



tput setaf 2
echo -e "\nCommit the change"
tput sgr0
git commit -a -m "Update version values and create changelog for release $VERSION"

tput setaf 2
echo -e "\nTag the release"
tput sgr0
git tag -a -m "Seed release $VERSION" $VERSION

tput setaf 2
echo -e "\nPush the changes"
tput sgr0
git push --tags

tput setaf 2
echo -e "\nSleeping 30 seconds so that tag is built prior to GitHub pages build"
tput sgr0
sleep 30

tput setaf 2
echo -e "\nCheckout back to ${DEFAULT_BRANCH}"
tput sgr0
git checkout ${DEFAULT_BRANCH}

tput setaf 2
echo -e "\nAdd $VERSION to version history"
tput sgr0
echo $VERSION >> .versions

tput setaf 2
echo -e "\nCommit the change"
tput sgr0
git commit -a -m "Add $VERSION to version history"

tput setaf 2
echo -e "\nPush the changes"
tput sgr0
git push

tput setaf 2
echo -e "\nDon't forget to update the release notes: https://github.com/ngageoint/seed/releases/edit/$VERSION"
tput sgr0
