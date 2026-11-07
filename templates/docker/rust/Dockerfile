# Multi-stage secure build for Rust
FROM rust:1.77-alpine AS builder

WORKDIR /src
RUN apk add --no-cache musl-dev
COPY Cargo.toml Cargo.lock* ./
COPY src ./src
RUN cargo build --release

# Distroless or minimal Alpine runtime
FROM alpine:3.19

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /src/target/release/app /app/server

USER appuser
EXPOSE 8080

ENTRYPOINT ["/app/server"]
