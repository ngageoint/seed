#!/bin/bash

# Capture command line arguments
INPUT=$1
OUTPUT=$2

echo ''
echo '----------------------------------------------------'
echo 'Calling algorithm with arguments ' $INPUT $OUTPUT
SCRIPT=my_alg.py

python $SCRIPT $INPUT $OUTPUT
rc=$?
echo 'Done calling algorithm - wrapper finished'
echo '----------------------------------------------------'
echo ''
exit $rc
