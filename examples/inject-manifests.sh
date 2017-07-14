#!/usr/bin/env bash

for DIRECTORY in $(ls -F | grep /)
do
    SEED_MANIFEST=$(jq -rc . < ${DIRECTORY}/seed.manifest.json | sed 's/\"/\\\"/g')
    sed -i "" -e "s/SEED_MANIFEST/"${SEED_MANIFEST}"/" ${DIRECTORY}/Dockerfile
done