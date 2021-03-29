FROM golang:1.16-alpine3.13

WORKDIR /work

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -o /go/bin/s3-copy
