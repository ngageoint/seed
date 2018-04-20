FROM bash

COPY app app
COPY . .

LABEL com.ngageoint.seed.manifest="SEED_MANIFEST"

RUN chmod -R 777 app

ENTRYPOINT ["/app/job.sh"]
