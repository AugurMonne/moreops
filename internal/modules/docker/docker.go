package docker

import (
	"context"
	"fmt"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

// Module implements Dockerfile, .dockerignore, and Docker Compose management.
type Module struct{}

func init() {
	plugin.Register(&Module{})
}

func (m *Module) Name() string {
	return "containers"
}

func (m *Module) Description() string {
	return "Manages Dockerfile, .dockerignore and Docker Compose configurations"
}

func (m *Module) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	if cfg.Modules.Containers == nil || !cfg.Modules.Containers.Enabled {
		return nil, nil
	}

	res := &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: make([]diff.FileChange, 0),
		Summary:     "Containers configuration",
	}

	port := 8080
	if cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Port > 0 {
		port = cfg.Modules.Kubernetes.Port
	}

	// 1. Generate Multi-Stage Hardened Dockerfile
	var dockerfileContent string
	switch cfg.Project.Language {
	case "go":
		dockerfileContent = fmt.Sprintf(`# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o /bin/app ./cmd/...

# Final minimal secure runtime
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /bin/app /app/server
USER 65532:65532
EXPOSE %d
ENTRYPOINT ["/app/server"]
`, port)

	case "node":
		dockerfileContent = fmt.Sprintf(`# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build --if-present

# Runtime stage
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder --chown=appuser:appgroup /app ./
USER appuser
EXPOSE %d
CMD ["npm", "start"]
`, port)

	case "python":
		dockerfileContent = fmt.Sprintf(`FROM python:3.11-slim AS runtime
WORKDIR /app
ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN groupadd -r appgroup && useradd -r -g appgroup -u 10001 appuser
COPY requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY --chown=appuser:appgroup . .
USER appuser
EXPOSE %d
CMD ["python", "main.py"]
`, port)

	case "rust":
		dockerfileContent = fmt.Sprintf(`FROM rust:1.77-alpine AS builder
WORKDIR /src
RUN apk add --no-cache musl-dev
COPY Cargo.toml Cargo.lock* ./
COPY src ./src
RUN cargo build --release

FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /app
COPY --from=builder /src/target/release/app /app/server
EXPOSE %d
ENTRYPOINT ["/app/server"]
`, port)

	default:
		dockerfileContent = fmt.Sprintf(`FROM alpine:3.19
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /app
EXPOSE %d
CMD ["sh"]
`, port)
	}

	fcDoc, err := diff.ComputeFileChange("Dockerfile", dockerfileContent)
	if err != nil {
		return nil, err
	}
	res.FileChanges = append(res.FileChanges, fcDoc)

	// 2. Generate .dockerignore
	dockerignoreContent := `.git
.gitignore
.moreops
bin
dist
tmp
node_modules
vendor
*.log
.env*
.DS_Store
`
	fcIgn, err := diff.ComputeFileChange(".dockerignore", dockerignoreContent)
	if err != nil {
		return nil, err
	}
	res.FileChanges = append(res.FileChanges, fcIgn)

	// 3. Compose file if enabled
	if cfg.Modules.Containers.Compose {
		composeContent := fmt.Sprintf(`version: "3.8"

services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    image: %s:dev
    ports:
      - "%d:%d"
    restart: unless-stopped
    environment:
      - PORT=%d
      - ENV=development
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "%d"]
      interval: 10s
      timeout: 5s
      retries: 3
`, cfg.Project.Name, port, port, port, port)

		fcComp, err := diff.ComputeFileChange("docker-compose.yml", composeContent)
		if err != nil {
			return nil, err
		}
		res.FileChanges = append(res.FileChanges, fcComp)
	}

	return res, nil
}

func (m *Module) Apply(ctx context.Context, plan *plugin.PlanResult) error {
	return plugin.ExecutePlan(plan)
}
