FROM buildpack-deps:jessie-scm
MAINTAINER Jingkai He

ENV SWIFT_PATH /swift

RUN mkdir -p $SWIFT_PATH && chmod -R 777 $SWIFT_PATH
WORKDIR $SWIFT_PATH

ENV VERSION swift-2.2-RELEASE
ENV PLATFORM ubuntu14.04

# Install Swift
RUN apt-get update -y
RUN apt-get install git cmake ninja-build \
  clang uuid-dev libicu-dev icu-devtools libbsd-dev \
  libedit-dev libxml2-dev libsqlite3-dev swig libpython-dev \
  libncurses5-dev pkg-config -y

RUN wget https://swift.org/builds/swift-2.2-release/ubuntu1404/swift-2.2-RELEASE/$VERSION-$PLATFORM.tar.gz
RUN tar xzf $VERSION-$PLATFORM.tar.gz

RUN rm $VERSION-$PLATFORM.tar.gz
RUN mv $VERSION-$PLATFORM/* $SWIFT_PATH && rm -rf $VERSION-$PLATFORM

ENV PATH $SWIFT_PATH/usr/bin:$PATH
# Tidy up
RUN apt-get clean

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

# RUN gpg --verify $VERSION-$PLATFORM.tar.gz.sig
ENTRYPOINT ["/entrypoint.sh"]
