#!/usr/bin/env bash

# Usage:

# ./set-version.sh 0.0.5
VERSION=$1

jq_in_place() {
    tmp=$(mktemp)
    jq $1 $2 > ${tmp}
    mv ${tmp} $2
}


# Update version placeholders
FILES=$(grep -r SEED_VERSION examples validator sections *.adoc | cut -d ':' -f 1 | sort | uniq)

for FILE in ${FILES}
do
    sed -i "" -e "s/SEED_VERSION/"$1"/g" ${FILE}
done

# Update schemas
jq_in_place .properties.seedVersion.pattern=\"${VERSION}\" schema/seed.manifest.schema.json
jq_in_place .properties.seedVersion.pattern=\"${VERSION}\" schema/seed.metadata.schema.json

# Update examples
FILES=$(grep -r seedVersion examples/ | cut -d ':' -f 1 | sort | uniq | grep -v Dockerfile)

for FILE in ${FILES}
do
    jq_in_place .seedVersion=\"${VERSION}\" ${FILE}
done
