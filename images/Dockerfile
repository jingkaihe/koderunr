FROM ubuntu
MAINTAINER Jingkai He

# Install Golang
RUN apt-get update -y
RUN apt-get install software-properties-common -y

RUN add-apt-repository ppa:ubuntu-lxc/lxd-stable -y
RUN apt-get update -y
RUN apt-get install golang -y

COPY entrypoint.sh /

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

RUN rbenv install 2.3.0
RUN rbenv global 2.3.0
RUN rbenv rehash
ENV PATH $RBENV_PATH/shims:$PATH

# Tidy up
RUN apt-get clean

# Entry point
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
