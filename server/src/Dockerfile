# syntax=docker/dockerfile:1

FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./
COPY cmd/ ./cmd/
COPY pkg/ ./pkg/

RUN go build -o ./timeasy-server ./cmd/service

EXPOSE 8080

CMD ["./timeasy-server"]
