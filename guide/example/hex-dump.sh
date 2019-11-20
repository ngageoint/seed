#!/usr/bin/env sh

## Usage:
## hex-dump.sh INPUT_FILE BYTE_COUNT OUTPUT_DUMP_FILE

INPUT_FILE=$1
BYTE_COUNT=$2
OUTPUT_DUMP_FILE=$3

echo "Invoked with command line: $*"

head -c $BYTE_COUNT $INPUT_FILE | od -x | tee $OUTPUT_DUMP_FILE

echo "Execution complete."
