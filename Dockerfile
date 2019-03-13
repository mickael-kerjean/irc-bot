FROM golang:alpine as builder

ENV GOPATH /go/

WORKDIR /app/
RUN apk add --update git make build-base && \
    mkdir -p /go/src/github.com/mickael-kerjean && \
    cd /go/src/github.com/mickael-kerjean && \
    git clone https://github.com/mickael-kerjean/prologic_cadmus_fork --depth 1 && \
    cd prologic_cadmus_fork && \
    go get -v ./... && \
    go build -o /app/run cmd/cadmus/main.go

FROM alpine:latest
COPY --from=builder /app/ /app/
RUN addgroup -S bot && adduser -S -g bot bot && \
    mkdir -p /app/logs/ && \
    chown -R bot:bot /app

USER bot
VOLUME ["/app/logs/"]
CMD ["/app/run", "irc.freenode.net:6667"]
