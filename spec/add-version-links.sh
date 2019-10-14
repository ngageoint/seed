#!/usr/bin/env bash

# Ensure script directory is CWD
pushd "${0%/*}"

for VERSION in $(cat ../.versions) master
do
    echo image:https://img.shields.io/badge/seed-${VERSION}-brightgreen.svg[link="${VERSION}/"] >> index.adoc
done

popd > /dev/null
