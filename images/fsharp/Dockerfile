FROM microsoft/dotnet:1.0.0-preview2-sdk
MAINTAINER Jingkai He


ENV FSHARP_PATH /fsharp
RUN mkdir -p $FSHARP_PATH

WORKDIR $FSHARP_PATH

COPY entrypoint.sh /
COPY project.json $FSHARP_PATH

RUN chmod +x /entrypoint.sh && \
  dotnet restore

ENTRYPOINT ["/entrypoint.sh"]
