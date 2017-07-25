FROM alpine

RUN adduser -D -g '' jtobe

ADD app app
ADD results_manifest.json results_manifest.json

RUN chmod -R 777 app
RUN chmod 777 results_manifest.json

ENTRYPOINT ["app/extract.sh"]

LABEL com.ngageoint.seed.manifest="{\"seedVersion\":\"0.0.5\",\"job\":{\"name\":\"extractor\",\"algorithmVersion\":\"0.1.0\",\"packageVersion\":\"0.1.0\",\"title\":\"Extractor\",\"description\":\"Reads a zip file and extracts the contents\",\"tag\":[\"zip\"],\"authorName\":\"John Doe\",\"authorEmail\":\"jdoe@example.com\",\"authorUrl\":\"http:\/\/www.example.com\",\"timeout\":3600,\"cpus\":10,\"mem\":10240,\"sharedMem\":0,\"storage\":0,\"interface\":{\"cmd\":\"\${ZIP} -d \${JOB_OUTPUT_DIR}\",\"inputData\":{\"files\":[{\"name\":\"ZIP\",\"required\":true,\"mediaType\":[\"application\/zip\"]}]},\"outputData\":{\"files\":[{\"name\":\"output_file_tiffs\",\"mediaType\":\"image\/png\",\"count\":\"2\",\"pattern\":\"s*.png\"}],\"json\":[{\"name\":\"Num Files\",\"key\":\"Num_Files\",\"type\":\"integer\"},{\"name\":\"Filenames\",\"type\":\"string\"},{\"name\":\"dummy\",\"type\":\"integer\",\"required\":false}]},\"mounts\":[{\"name\":\"MOUNTAIN\",\"path\":\"\/the\/mountain\",\"mode\":\"ro\"}],\"settings\":[{\"name\":\"HELLO\",\"secret\":false}]},\"errorMapping\":[{\"code\":1,\"title\":\"Error Name\",\"description\":\"Error Description\",\"category\":\"system\"},{\"code\":2,\"title\":\"Error Name\",\"description\":\"Error Description\",\"category\":\"data\"},{\"code\":3,\"title\":\"Error Name\",\"description\":\"Error Description\",\"category\":\"algorithm\"}]}}"
