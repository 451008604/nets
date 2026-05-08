FROM alpine:3.23

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY client/client .

ENV GODEBUG="netdns=cgo"
ENV GOTRACEBACK=crash

CMD ["/app/client"]
