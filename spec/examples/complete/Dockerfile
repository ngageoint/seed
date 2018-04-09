FROM bash

COPY app app
COPY . .

RUN chmod -R 777 app

ENTRYPOINT ["/app/job.sh"]

LABEL com.ngageoint.seed.manifest="SEED_MANIFEST"