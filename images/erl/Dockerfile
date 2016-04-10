FROM ubuntu

# Install Erlang and Elixir
RUN apt-get install wget -y
RUN wget https://packages.erlang-solutions.com/erlang-solutions_1.0_all.deb
RUN dpkg -i erlang-solutions_1.0_all.deb
RUN apt-get update -y
RUN apt-get install esl-erlang -y
RUN apt-get install elixir -y

# locale settings
RUN locale-gen "en_US.UTF-8"
ENV LC_ALL en_US.UTF-8

# Tidy up
RUN apt-get clean

# Entry point
COPY entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
