FROM alpine:3.4

MAINTAINER Jingkai He

RUN apk add --no-cache g++

ENV C_PATH /c
RUN mkdir -p $C_PATH
WORKDIR $C_PATH

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
