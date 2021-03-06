#!/usr/bin/env bash

# Ensure script directory is CWD
pushd "${0%/*}"

# Inject the manifest for each example into SEED_MANIFEST placeholder
for DIRECTORY in $(ls -F | grep /)
do
    SEED_MANIFEST=$(jq -rc . < ${DIRECTORY}/seed.manifest.json | sed 's^\"^\\\\\"^g' | sed 's^\$^\\\\\$^g' | sed 's^/^\\\\/^g')
    sed -i.bak "s^SEED_MANIFEST^${SEED_MANIFEST}^" ${DIRECTORY}/Dockerfile
done

popd > /dev/null
