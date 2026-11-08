package docker

import (
	"context"
	"fmt"
	"strings"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
	"moreops/internal/core/template"
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
		defaultGo := fmt.Sprintf(`# Build stage
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
		dockerfileContent = template.Resolve("docker/go", defaultGo)

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
		dockerfileContent = fmt.Sprintf(`FROM rust:1.78-alpine AS builder
WORKDIR /src
RUN apk add --no-cache musl-dev
COPY Cargo.toml Cargo.lock* ./
COPY src ./src
RUN cargo build --release

FROM alpine:3.20
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
WORKDIR /app
COPY --from=builder /src/target/release/app /app/server
EXPOSE %d
ENTRYPOINT ["/app/server"]
`, port)

	default:
		dockerfileContent = fmt.Sprintf(`FROM alpine:3.20
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
		var extraEnv strings.Builder
		var dependsOn strings.Builder
		var extraServices strings.Builder
		var volumes strings.Builder

		if cfg.Modules.Database != nil && cfg.Modules.Database.Enabled {
			switch cfg.Modules.Database.Engine {
			case "mysql":
				extraEnv.WriteString("      - DB_HOST=mysql\n      - DB_PORT=3306\n      - DB_NAME=dev_db\n      - DB_USER=dev\n      - DB_PASSWORD=dev\n")
				dependsOn.WriteString("    depends_on:\n      mysql:\n        condition: service_healthy\n")
				extraServices.WriteString(`
  mysql:
    image: mysql:8.0
    restart: unless-stopped
    ports:
      - "3306:3306"
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: dev_db
      MYSQL_USER: dev
      MYSQL_PASSWORD: dev
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - mysqldata:/var/lib/mysql
`)
				volumes.WriteString("\nvolumes:\n  mysqldata:\n")
			case "redis":
				extraEnv.WriteString("      - REDIS_HOST=redis\n      - REDIS_PORT=6379\n")
				dependsOn.WriteString("    depends_on:\n      redis:\n        condition: service_healthy\n")
				extraServices.WriteString(`
  redis:
    image: redis:7-alpine
    restart: unless-stopped
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 3s
      retries: 5
    volumes:
      - redisdata:/data
`)
				volumes.WriteString("\nvolumes:\n  redisdata:\n")
			default: // postgres
				extraEnv.WriteString("      - DB_HOST=postgres\n      - DB_PORT=5432\n      - DB_NAME=dev_db\n      - DB_USER=dev\n      - DB_PASSWORD=dev\n")
				dependsOn.WriteString("    depends_on:\n      postgres:\n        condition: service_healthy\n")
				extraServices.WriteString(`
  postgres:
    image: postgres:16-alpine
    restart: unless-stopped
    ports:
      - "5432:5432"
    environment:
      POSTGRES_DB: dev_db
      POSTGRES_USER: dev
      POSTGRES_PASSWORD: dev
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U dev -d dev_db"]
      interval: 5s
      timeout: 5s
      retries: 5
    volumes:
      - pgdata:/var/lib/postgresql/data
`)
				volumes.WriteString("\nvolumes:\n  pgdata:\n")
			}
		}

		if cfg.Modules.Messaging != nil && cfg.Modules.Messaging.Enabled {
			extraEnv.WriteString("      - RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/\n      - RABBITMQ_HOST=rabbitmq\n      - RABBITMQ_PORT=5672\n")
			if dependsOn.Len() == 0 {
				dependsOn.WriteString("    depends_on:\n")
			}
			dependsOn.WriteString("      rabbitmq:\n        condition: service_healthy\n")
			extraServices.WriteString(`
  rabbitmq:
    image: rabbitmq:3.13-management-alpine
    restart: unless-stopped
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "-q", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    volumes:
      - rabbitmqdata:/var/lib/rabbitmq
`)
			if volumes.Len() == 0 {
				volumes.WriteString("\nvolumes:\n")
			}
			volumes.WriteString("  rabbitmqdata:\n")
		}

		defaultCompose := fmt.Sprintf(`version: "3.8"

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
%s%s    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "%d"]
      interval: 10s
      timeout: 5s
      retries: 3
%s%s`, cfg.Project.Name, port, port, port, extraEnv.String(), dependsOn.String(), port, extraServices.String(), volumes.String())

		composeContent := template.Resolve("docker/compose", defaultCompose)

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
