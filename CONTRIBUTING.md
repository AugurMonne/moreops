# Contributing to MoreOps

Thank you for your interest in contributing to MoreOps! We welcome pull requests, bug reports, feature suggestions, and documentation improvements.

---

## Development Setup

### Prerequisites
- Go 1.22+
- Docker
- Git
- Optional: `k3d`, `kubectl`, `helm`, `golangci-lint`

### Building from Source

```bash
git clone https://github.com/moreops/moreops.git
cd moreops
make build
```

Run tests to ensure everything passes:

```bash
make test
```

Run linter:

```bash
make lint
```

---

## Coding Guidelines

- All Go code must be formatted using standard `gofmt` and pass `golangci-lint`.
- Every new module or command must include unit tests (`*_test.go`).
- Maintain backward compatibility of `moreops.yaml` schemas.
- Preserve existing comments and docstrings.
- Adhere to the core principle: **Orchestrate existing tools rather than reinventing them**.

---

## Submitting Pull Requests

1. Fork the repository and create a new branch:
   ```bash
   git checkout -b feature/my-awesome-feature
   ```
2. Commit your changes using conventional commit messages:
   - `feat: add support for new storage provider`
   - `fix: resolve selector mismatch in k8s generator`
   - `docs: update CLI reference`
3. Ensure all tests and linters pass (`make test && make lint`).
4. Push your branch and open a Pull Request against `main`.
