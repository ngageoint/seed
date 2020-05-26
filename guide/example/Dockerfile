FROM python:3.8-alpine

RUN apk add build-base python-dev py-pip jpeg-dev zlib-dev
ENV LIBRARY_PATH=/lib:/usr/lib

RUN pip install --upgrade pip && \
    pip install --upgrade Pillow >= 7

WORKDIR /algo

ADD ./image_rotate.py .
ADD ./image_rotate.sh .