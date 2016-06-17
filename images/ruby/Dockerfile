FROM buildpack-deps:jessie-scm
MAINTAINER Jingkai He

RUN apt-get update -y
RUN apt-get install software-properties-common -y

# Install Ruby
RUN apt-get install python-software-properties -y
RUN apt-get install build-essential git-core wget -y
RUN apt-get install libssl-dev libreadline-dev zlib1g-dev -y

ENV RBENV_PATH /usr/local/rbenv
RUN git clone git://github.com/sstephenson/rbenv.git $RBENV_PATH
RUN git clone git://github.com/sstephenson/ruby-build.git $RBENV_PATH/plugins/ruby-build
RUN $RBENV_PATH/plugins/ruby-build/install.sh
ENV PATH $RBENV_PATH/bin:$PATH
RUN echo 'eval "$(rbenv init -)"' >> /etc/profile.d/rbenv.sh

ENV CONFIGURE_OPTS --disable-install-doc
RUN echo 'gem: --no-rdoc --no-ri' >> /.gemrc

RUN apt-get install -y --no-install-recommends openjdk-7-jre-headless

# Install MRIs
ADD ./versions.txt /root/versions.txt
RUN xargs -L 1 rbenv install < /root/versions.txt

RUN rbenv global 2.3.0
RUN rbenv rehash
ENV PATH /root/.rbenv/shims:$PATH

# Tidy up
RUN apt-get clean

COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
