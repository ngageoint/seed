#!/bin/sh

INPUT=$1
DEGREES=$2
OUTPUT_DIR=$3

echo ''
echo '-----------------------------------------------------------------'
echo 'Rotating image with arguments '${INPUT} ${DEGREES} ${OUTPUT_DIR}

python ./image_rotate.py $INPUT $DEGREES $OUTPUT_DIR
rc=$?

echo 'Done rotating image'
echo '-----------------------------------------------------------------'
echo ''
exit ${rc}