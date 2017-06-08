#!/usr/bin/env bash

# Ensure current directory is the one containing this script
cd `dirname $0` 

docker build ../../examples/complete/ -t my-algorithm-0.1.0-seed:0.1.0
docker build ../../examples/random-number/ -t random-number-gen-0.1.0-seed:0.1.0
docker build ../../examples/watermark/ -t image-watermark-0.1.0-seed:0.1.0
docker build ../../examples/watermark/ -t seed-test/watermark

docker build invalid-missing-job/ -t seed-test/invalid-missing-job
docker build invalid-missing-job-interface-inputdata-files-name/ -t missing-filename-0.1.0-seed:0.1.0
