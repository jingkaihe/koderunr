
FROM ubuntu
MAINTAINER Jingkai He

RUN sudo apt-get update -y
RUN sudo apt-get install software-properties-common -y

RUN sudo add-apt-repository ppa:ubuntu-lxc/lxd-stable -y
RUN sudo apt-get update -y
RUN sudo apt-get install golang -y

COPY entrypoint.sh /

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
