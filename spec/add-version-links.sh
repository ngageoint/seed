#!/usr/bin/env bash

# Ensure script directory is CWD
cd "${0%/*}"

for VERSION in $(cat ../.versions)
do
    echo image:https://img.shields.io/badge/seed-${VERSION}-brightgreen.svg[link="https://ngageoint.github.io/seed/seed-${VERSION}.html"] >> index.adoc
done

cd - > /dev/null