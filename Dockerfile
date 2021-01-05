FROM golang:1.12.0-alpine3.9

ENV GO111MODULE=on

RUN mkdir /app
ADD . /app

COPY  go.mod /app
COPY go.sum /app
RUN go mod download


WORKDIR /app
RUN go build -o main .
CMD ["/app/main"]