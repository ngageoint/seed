#!/usr/bin/env bash
set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

# Validate each seed.manifest.json
for DIRECTORY in $(ls -F | grep /)
do
    echo "Validating manifest for example in ${DIRECTORY}..."
    ajv validate -s ../schema/seed.manifest.schema.json -d ${DIRECTORY}seed.manifest.json
    for METADATA in $(ls ${DIRECTORY} | grep '.metadata.json')
    do
        echo "Validating ${DIRECTORY}${METADATA} against schema..."
        ajv -s ../schema/seed.metadata.schema.json -d ${DIRECTORY}${METADATA} --missing-refs=ignore
    done
done

popd > /dev/null

