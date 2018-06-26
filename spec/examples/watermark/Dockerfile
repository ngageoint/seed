FROM alpine

ENTRYPOINT /app/watermark.py

LABEL com.ngageoint.seed.manifest="{\"seedVersion\":\"1.0.0\",\"job\":{\"name\":\"image-watermark\",\"jobVersion\":\"0.1.0\",\"packageVersion\":\"0.1.0\",\"title\":\"Image Watermarker\",\"description\":\"Processes an input PNG and outputs watermarked PNG.\",\"maintainer\":{\"name\":\"John Doe\",\"email\":\"jdoe@example.com\"},\"timeout\":30,\"interface\":{\"command\":\"\${INPUT_IMAGE} \${OUTPUT_DIR}\",\"inputs\":{\"files\":[{\"name\":\"INPUT_IMAGE\"}]},\"outputs\":{\"files\":[{\"name\":\"OUTPUT_IMAGE\",\"pattern\":\"*_watermark.png\"}]}},\"resources\":{\"scalar\":[{\"name\":\"cpus\",\"value\":1},{\"name\":\"mem\",\"value\":64}]},\"errors\":[{\"code\":1,\"name\":\"image-Corrupt\",\"description\":\"Image input is not recognized as a valid PNG.\",\"category\":\"data\"},{\"code\":2,\"name\":\"algorithm-failure\"}]}}"
