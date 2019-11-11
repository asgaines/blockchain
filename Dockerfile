FROM golang:1.13

WORKDIR /go/src/github.com/asgaines/blockchain

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /go/bin/blockchain .

EXPOSE 20403

ENTRYPOINT ["/go/bin/blockchain"]

CMD ["-help"]