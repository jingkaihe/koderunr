FROM microsoft/dotnet:1.0.0-preview2-sdk
MAINTAINER Jingkai He


ENV DOTNET_PATH /dotnet
RUN mkdir -p $DOTNET_PATH

WORKDIR $DOTNET_PATH

COPY entrypoint.sh /
COPY project.json $DOTNET_PATH

RUN chmod +x /entrypoint.sh && \
  dotnet restore

ENTRYPOINT ["/entrypoint.sh"]
