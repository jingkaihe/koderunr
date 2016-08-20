FROM python:3.4.5-alpine
MAINTAINER Jingkai He

ENV PYTHON_PATH /python
RUN mkdir -p $PYTHON_PATH
WORKDIR $PYTHON_PATH

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

