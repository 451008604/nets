FROM debian:stable-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY server/server .

CMD ["/app/server"]
