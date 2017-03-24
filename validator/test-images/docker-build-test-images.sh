#!/usr/bin/env bash

# Ensure current directory is the one containing this script
cd `dirname $0` 

docker build ../../examples/complete/ -t seed-test/complete
docker build ../../examples/random-number/ -t seed-test/random-number
docker build ../../examples/watermark/ -t seed-test/watermark

docker build invalid-missing-jobs/ -t seed-test/invalid-missing-jobs
docker build invalid-missing-job-interface-inputdata-files-name/ -t seed-test/invalid-missing-job-interface-inputdata-files-name
