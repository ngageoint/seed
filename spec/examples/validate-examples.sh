#!/usr/bin/env bash
set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

# Validate each seed.manifest.json
for DIRECTORY in $(ls -F | grep /)
do
    echo "Validating manifest for example in ${DIRECTORY}..."
    json validate --schema-file=../schema/seed.manifest.schema.json < ${DIRECTORY}/seed.manifest.json
done

popd > /dev/null

