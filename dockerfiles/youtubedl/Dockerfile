FROM golang:latest

WORKDIR /app

ADD client .
RUN go build -o client main.go

CMD ./client
