FROM alpine

RUN adduser -D -g '' jtobe

ADD app app
ADD results_manifest.json results_manifest.json
ADD seed.png.metadata.json seed.png.metadata.json

RUN chmod -R 777 app
RUN chmod 777 results_manifest.json
RUN chmod 777 seed.png.metadata.json

ENTRYPOINT ["app/extract.sh"]

LABEL com.ngageoint.seed.manifest="SEED_MANIFEST"
