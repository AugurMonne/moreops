# MoreOps Manual Testing Guide

A step-by-step walkthrough to test all MoreOps capabilities from your terminal.

---

## Step 0: Build & Environment Check

1. Build the binary:
   ```bash
   make build
   ```
2. Check installed dependencies:
   ```bash
   ./bin/moreops doctor
   ```

---

## Step 1: Create a Demo Service & Run `moreops init`

1. Create a temporary test service:
   ```bash
   mkdir -p /tmp/demo-service && cd /tmp/demo-service

   cat << 'EOF' > main.go
   package main
   import "net/http"
   func main() {
       http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
       http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("# HELP test 1\n")) })
       http.ListenAndServe(":8080", nil)
   }
   EOF

   go mod init demo-service
   ```

2. Run the wizard:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops init
   ```
   - Notice auto-detection: `🔍 Detected: GO`.
   - Press `Enter` to accept defaults (or select components like `1 2 3 4 5 6` or `14`).
   - You can specify initial replica count (default: 2) and enable HPA with max replicas (default: 10).
   - A valid `moreops.yaml` configuration is created.

---

## Step 2: Preview Changes (`moreops plan`)

```bash
/home/augur/Documents/forgeops/bin/moreops plan
```
- Previews the unified diff without writing any files to disk.
- Summarizes planned creations.

---

## Step 3: Apply Changes (`moreops apply`)

1. Apply with confirmation:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops apply
   ```
   Type `yes` when prompted.
2. Verify files are generated: `Dockerfile`, `.dockerignore`, `charts/demo-service/`, etc.
3. Re-run plan:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops plan
   ```
   *Expected output*: `✓ No infrastructure changes planned for environment [dev]. All resources are up to date.`

---

## Step 4: Status Check (`moreops status`)

```bash
/home/augur/Documents/forgeops/bin/moreops status
```
Shows project metadata, active environments, enabled modules, and existing manifests on disk.

---

## Step 5: Static Analysis & Security Linter (`moreops validate`)

1. Run full validation:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops validate
   ```
2. Selective checks:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops validate --security
   /home/augur/Documents/forgeops/bin/moreops validate --rbac
   ```

---

## Step 6: Environment Management (`moreops env`)

1. List environments:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env list
   ```
2. Add a new environment:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env add staging --cluster k8s-staging --namespace staging
   ```
3. Diff environments:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env diff dev staging
   ```
4. Switch active environment:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env switch staging
   ```

---

## Step 7: Helm Chart Packaging (`moreops chart`)

Package the Helm chart (uses native pure Go packager if Helm is not installed):
```bash
/home/augur/Documents/forgeops/bin/moreops chart package charts/demo-service
```
Generates `demo-service-0.1.0.tgz`.

---

## Step 8: Release Automation (`moreops release`)

Bump minor version:
```bash
/home/augur/Documents/forgeops/bin/moreops release minor
```
Updates `moreops.yaml`, `Chart.yaml`, and appends release notes to `CHANGELOG.md`.

---

## Step 9: Cleanup

```bash
cd /home/augur/Documents/forgeops
rm -rf /tmp/demo-service
```
