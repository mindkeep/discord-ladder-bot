# Build stage
FROM rust:latest AS build

WORKDIR /build

COPY Cargo.toml Cargo.lock src ./

RUN cargo build --release

# Final stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=build /build/target/release/discord-ladder-bot .

CMD ["./discord-ladder-bot"]