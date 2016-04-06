FROM ubuntu
MAINTAINER Jingkai He

# Install Golang
RUN apt-get update -y
RUN apt-get install software-properties-common -y

RUN add-apt-repository ppa:ubuntu-lxc/lxd-stable -y
RUN apt-get update -y
RUN apt-get install golang -y

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

# Install MRIs
RUN rbenv install 2.3.0
RUN rbenv install 1.9.3-p550

RUN rbenv global 2.3.0
RUN rbenv rehash
ENV PATH /root/.rbenv/shims:$PATH

# Install Erlang and Elixir
RUN wget https://packages.erlang-solutions.com/erlang-solutions_1.0_all.deb
RUN dpkg -i erlang-solutions_1.0_all.deb
RUN apt-get update
RUN apt-get install esl-erlang -y
RUN apt-get install elixir -y

# Tidy up
RUN apt-get clean
RUN locale-gen "en_US.UTF-8"
ENV LC_ALL en_US.UTF-8

# Entry point
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
