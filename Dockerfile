#FROM ubuntu:latest
FROM golang:latest
LABEL authors="ekaterina"

ENV GOPATH=/

COPY ./ ./

RUN go build -o cmd/app .

ENTRYPOINT ["./cmd/app"]