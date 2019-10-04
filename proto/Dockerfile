FROM golang:1.13

RUN apt-get update && \
    apt-get -y install \
      git unzip wget \
      build-essential autoconf libtool \
      default-jre-headless

# Go setup
RUN go get google.golang.org/grpc
RUN go get google.golang.org/genproto/...
RUN go get github.com/golang/protobuf/protoc-gen-go
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
RUN go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger
RUN go get github.com/fiorix/protoc-gen-cobra
RUN go get github.com/golang/mock/mockgen

# Prototool
ARG PROTOTOOL_VER=1.8.0
RUN curl -sSL https://github.com/uber/prototool/releases/download/v${PROTOTOOL_VER}/prototool-$(uname -s)-$(uname -m) \
      -o /usr/local/bin/prototool && \
      chmod a+rx /usr/local/bin/prototool

# Protoc
ARG PROTOC_VER=3.9.1
RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v${PROTOC_VER}/protoc-${PROTOC_VER}-linux-x86_64.zip -O /tmp/protoc-bin.zip \
    && unzip /tmp/protoc-bin.zip -d /usr/local/

# Java gRPC
ARG PROTOC_JAVA_GRPC=1.23.0
RUN curl -fsSL http://central.maven.org/maven2/io/grpc/protoc-gen-grpc-java/${PROTOC_JAVA_GRPC}/protoc-gen-grpc-java-${PROTOC_JAVA_GRPC}-linux-x86_64.exe -o /usr/local/bin/protoc-gen-grpc-java \
    && chmod a+rx /usr/local/bin/protoc-gen-grpc-java

# Scala
ARG SCALAPBC_VER=0.9.0-M7
RUN curl -fsSL "https://github.com/scalapb/ScalaPB/releases/download/v${SCALAPBC_VER}/scalapbc-${SCALAPBC_VER}.zip" -o /usr/local/bin/scalapbc.zip \
    && unzip /usr/local/bin/scalapbc.zip -d /usr/local/bin \
    && ln -s "/usr/local/bin/scalapbc-${SCALAPBC_VER}/bin/scalapbc" /usr/local/bin/scalapbc

WORKDIR /go/src/github.com/asgaines/blockchain
CMD ["/bin/bash", "./proto/build.sh"]
