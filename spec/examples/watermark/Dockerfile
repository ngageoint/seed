FROM alpine

ENTRYPOINT /app/watermark.py

LABEL com.ngageoint.seed.manifest="{\"seedVersion\":\"0.0.6\",\"job\":{\"name\":\"image-watermark\",\"algorithmVersion\":\"0.1.0\",\"packageVersion\":\"0.1.0\",\"title\":\"Image Watermarker\",\"description\":\"Processes an input PNG and outputs watermarked PNG.\",\"authorName\":\"John Doe\",\"timeout\":30,\"interface\":{\"cmd\":\"\${INPUT_IMAGE} \${OUTPUT_DIR}\",\"resources\":{\"scalar\":[{\"name\":\"cpu\",\"value\":1},{\"name\":\"mem\",\"value\":64}]},\"inputData\":{\"files\":[{\"name\":\"INPUT_IMAGE\",\"mediaType\":[\"image\/png\"]}]},\"outputData\":{\"files\":[{\"name\":\"OUTPUT_IMAGE\",\"mediaType\":\"image\/png\",\"pattern\":\"*_watermark.png\"}]}},\"errorMapping\":[{\"code\":1,\"title\":\"Image Corrupt\",\"description\":\"Image input is not recognized as a valid PNG.\",\"category\":\"data\"}]}}"
