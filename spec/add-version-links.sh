#!/usr/bin/env bash

# Ensure script directory is CWD
pushd "${0%/*}"

for VERSION in $(cat ../.versions)
do
    echo image:https://img.shields.io/badge/seed-${VERSION}-brightgreen.svg[link="https://ngageoint.github.io/seed/${VERSION}/seed.html"] >> index.adoc
done

popd > /dev/null