#! /bin/sh
unzip $*
cp results_manifest.json $3
cat results_manifest.json
cp seed.png.metadata.json $3
cat seed.png.metadata.json
ls -lR /the
echo $HELLO
