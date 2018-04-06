#!/usr/bin/env bash

err_count=0

version=1

if [ "$#" -ne 3 ]; then
    echo "Expected 3 input parameters: \${INPUT_FILE} \${OUTPUT_DIR} \${VERSION}"
    echo "Received" $*
    err_count=$((err_count+1))
else
    version=$3
fi

if [[ $3 != 1 && $3 != 2 ]]; then
    echo "Expected 1 or 2 for \${VERSION}, received " $3
    err_count=$((err_count+1))
fi

if [[ ! -r $1 ]]; then
    echo "Input file " $1 " does not exist or is not readable!"
    err_count=$((err_count+1))
fi

if [[ ! -d $2 ]]; then
    echo "Output directory " $2 " does not exist!"
    err_count=$((err_count+1))
fi

if [[ ! -w $2 ]]; then
    echo "Output directory " $2 " is not writeable!"
    err_count=$((err_count+1))
else
    cp *.png $2
    if [[ $3 == 1 ]]; then
        cp seed.outputs.json $2
        cat seed.outputs.json
        cp *.csv $2
    fi
    if [[ $3 == 2 ]]; then
        cp seed.outputs2.json $2
        cat seed.outputs2.json
    fi
fi

if [[ -z "${INPUT_FILE}" ]]; then
  echo "Need to set input INPUT_FILE to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [[ -z "${INPUT_JSON}" ]]; then
  echo "Need to set input INPUT_JSON to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [[ -z "${OUTPUT_DIR}" ]]; then
  echo "Need to set OUTPUT_DIR to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [[ ! -r "/the/container/path" ]]; then
    echo "Mounted directory /the/container/path specified by MOUNT_PATH does not exist or is not readable!"
    err_count=$((err_count+1))
fi

if [[ -w "/the/container/path" ]]; then
    echo "Mounted directory /the/container/path specified by MOUNT_PATH is writeable when ro was specified!"
    err_count=$((err_count+1))
fi

if [[ ! -w "/write" ]]; then
    echo "Mounted directory /write specified by WRITE_PATH is not writeable!"
    err_count=$((err_count+1))
fi

if [[ -z "${VERSION}" ]]; then
  echo "Need to set setting VERSION to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [[ -z "${DB_HOST}" ]]; then
  echo "Need to set setting DB_HOST to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [[ -z "${DB_PASS}" ]]; then
  echo "Need to set setting DB_PASS to non-empty environment variable"
  err_count=$((err_count+1))
fi

if [ "${CPUS}" != "10.0" ]; then
  echo "Need to set setting CPUS to environment variable with a value of 10.0"
  err_count=$((err_count+1))
fi

#TODO: add check for avaialble memory
if [ "${MEM}" != "1024.0" ]; then
  echo "Need to set setting MEM to environment variable with a value of 1024.0"
  err_count=$((err_count+1))
fi

if [ "${SHAREDMEM}" != "1024.0" ]; then
  echo "Need to set setting SHAREDMEM to environment variable with a value of 1024.0"
  err_count=$((err_count+1))
fi

#TODO: add check for avaialble disk space and calculate the required space based on input file size
if [[ -z "${DISK}" ]]; then
  echo "Need to set setting DISK to non-empty environment variable"
  err_count=$((err_count+1))
fi

echo "Encountered " $err_count " errors running test seed image"

exit $err_count