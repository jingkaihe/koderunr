FROM ubuntu
MAINTAINER Jingkai He

# Install Golang
RUN apt-get update -y
RUN apt-get install software-properties-common -y

RUN add-apt-repository ppa:ubuntu-lxc/lxd-stable -y
RUN apt-get update -y
RUN apt-get install golang -y


ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"
WORKDIR $GOPATH/src

# Tidy up
RUN apt-get clean

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
