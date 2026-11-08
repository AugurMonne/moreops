# Multi-stage secure build for Go applications
FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum* ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o /bin/app ./cmd/...

# Minimal distroless runtime
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=builder /bin/app /app/server

USER 65532:65532
EXPOSE 8080

ENTRYPOINT ["/app/server"]
