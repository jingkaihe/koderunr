FROM buildpack-deps:jessie-scm
MAINTAINER Jingkai He

# Install Golang
RUN apt-get update -y
RUN apt-get install gcc -y

# Tidy up
RUN apt-get clean

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
