FROM alpine:3.23

RUN apk add --no-cache ca-certificates wget

WORKDIR /app

COPY server/server .

ENV GODEBUG="netdns=cgo"
ENV GOTRACEBACK=crash

CMD ["/app/server"]
