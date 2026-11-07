# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /bin/moreops ./cmd/moreops

# Final Stage
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /workspace
COPY --from=builder /bin/moreops /usr/local/bin/moreops

USER 65532:65532

ENTRYPOINT ["/usr/local/bin/moreops"]
CMD ["--help"]
