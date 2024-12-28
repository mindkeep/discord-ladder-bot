FROM alpine:3.21

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the Go binary from the local build context
COPY bin/discord-ladder-bot .

CMD ["./discord-ladder-bot"]
