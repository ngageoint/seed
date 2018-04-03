FROM alpine

ADD app app
ADD *.png
ADD *.csv
ADD seed.outputs*.json

RUN chmod -R 777 app

ENTRYPOINT /app/job.sh

LABEL com.ngageoint.seed.manifest="SEED_MANIFEST"
