#!/usr/bin/env bash

# Ensure script directory is CWD
cd "${0%/*}"

# Inject the manifest for each example into SEED_MANIFEST placeholder
for DIRECTORY in $(ls -F | grep /)
do
    SEED_MANIFEST=$(jq -rc . < ${DIRECTORY}/seed.manifest.json | sed 's/\"/\\\\\"/g')
    sed -i "" -e "s^SEED_MANIFEST^${SEED_MANIFEST}^" ${DIRECTORY}/Dockerfile
done

cd -
