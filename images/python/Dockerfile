FROM buildpack-deps:jessie-scm
MAINTAINER Jingkai He

RUN apt-get update -y
RUN apt-get install software-properties-common -y

# Install Ruby
RUN apt-get install python-software-properties -y
RUN apt-get install build-essential git-core wget -y
RUN apt-get install libssl-dev libreadline-dev zlib1g-dev -y

RUN curl -L https://raw.githubusercontent.com/yyuu/pyenv-installer/master/bin/pyenv-installer | bash

ENV PYENV_ROOT /root/.pyenv
ENV PATH $PYENV_ROOT/shims:$PYENV_ROOT/bin:$PATH

ADD ./versions.txt /root/versions.txt
RUN xargs -L 1 pyenv install < /root/versions.txt

RUN pyenv global 2.7.6
RUN pyenv rehash

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
