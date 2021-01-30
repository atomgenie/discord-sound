FROM golang:latest

WORKDIR /app
LABEL org.opencontainers.image.source https://github.com/atomgenie/discord-sound

ADD client .
RUN go build -o client main.go

CMD ./client youtubedl
