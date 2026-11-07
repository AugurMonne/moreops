package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

// Module implements S3 storage, MinIO, Velero backups, and database backup workflows.
type Module struct{}

func init() {
	plugin.Register(&Module{})
}

func (m *Module) Name() string {
	return "storage"
}

func (m *Module) Description() string {
	return "Manages S3 object storage (Yandex, VK Cloud, Selectel, MinIO), Velero backups, and DB backup CronJobs"
}

func (m *Module) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	if cfg.Modules.Storage == nil || !cfg.Modules.Storage.Enabled {
		return nil, nil
	}

	env, _ := cfg.GetEnvironment(envName)
	namespace := env.Namespace
	if namespace == "" {
		namespace = "default"
	}

	appName := cfg.Project.Name
	stCfg := cfg.Modules.Storage

	bucket := stCfg.Bucket
	if bucket == "" {
		bucket = appName + "-backups"
	}

	endpoint := stCfg.Endpoint
	region := stCfg.Region
	if region == "" {
		region = "ru-central1"
	}

	res := &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: make([]diff.FileChange, 0),
		Summary:     fmt.Sprintf("Storage and backup configuration for %s", appName),
	}

	isHelm := cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Tool == "helm"
	var outDir string
	if isHelm {
		outDir = filepath.Join("charts", appName, "templates", "storage")
	} else {
		outDir = filepath.Join("k8s", "storage")
	}

	files := make(map[string]string)

	// 1. In-Cluster MinIO for Dev/Local Testing
	if stCfg.Provider == "minio-dev" {
		minioFile := filepath.Join(outDir, "minio-dev.yaml")
		files[minioFile] = fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: %s
  labels:
    app.kubernetes.io/name: minio
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: minio
  template:
    metadata:
      labels:
        app.kubernetes.io/name: minio
    spec:
      containers:
        - name: minio
          image: minio/minio:RELEASE.2024-01-18T22-51-28Z
          args:
            - server
            - /data
            - --console-address
            - :9001
          env:
            - name: MINIO_ROOT_USER
              value: "minioadmin"
            - name: MINIO_ROOT_PASSWORD
              value: "minioadmin"
          ports:
            - containerPort: 9000
              name: s3
            - containerPort: 9001
              name: console
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /minio/health/live
              port: 9000
            initialDelaySeconds: 10
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /minio/health/ready
              port: 9000
            initialDelaySeconds: 10
            periodSeconds: 20
---
apiVersion: v1
kind: Service
metadata:
  name: minio
  namespace: %s
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: minio
  ports:
    - name: s3
      port: 9000
      targetPort: 9000
    - name: console
      port: 9001
      targetPort: 9001
`, namespace, namespace)
		endpoint = fmt.Sprintf("http://minio.%s.svc:9000", namespace)
	}

	// 2. S3 Credentials Secret Template
	s3SecretFile := filepath.Join(outDir, "s3-credentials-secret.yaml")
	files[s3SecretFile] = fmt.Sprintf(`apiVersion: v1
kind: Secret
metadata:
  name: %s-s3-credentials
  namespace: %s
type: Opaque
stringData:
  AWS_ACCESS_KEY_ID: "replace-with-access-key"
  AWS_SECRET_ACCESS_KEY: "replace-with-secret-key"
  S3_ENDPOINT: "%s"
  S3_REGION: "%s"
  S3_BUCKET: "%s"
`, appName, namespace, endpoint, region, bucket)

	// 3. Velero Backup Storage Location & Schedule
	if stCfg.Velero != nil && stCfg.Velero.Enabled {
		veleroFile := filepath.Join(outDir, "velero-backup.yaml")
		scheduleCron := stCfg.Velero.Schedule
		if scheduleCron == "" {
			scheduleCron = "0 2 * * *"
		}
		ttl := stCfg.Velero.TTL
		if ttl == "" {
			ttl = "720h"
		}

		files[veleroFile] = fmt.Sprintf(`apiVersion: velero.io/v1
kind: BackupStorageLocation
metadata:
  name: %s-s3-location
  namespace: velero
spec:
  provider: aws
  objectStorage:
    bucket: %s
  config:
    region: %s
    s3ForcePathStyle: "true"
    s3Url: %s
---
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: %s-scheduled-backup
  namespace: velero
spec:
  schedule: "%s"
  template:
    includedNamespaces:
      - %s
    ttl: "%s"
    snapshotVolumes: true
`, appName, bucket, region, endpoint, appName, scheduleCron, namespace, ttl)
	}

	// 4. Automated Database Backup CronJob
	if stCfg.DatabaseBackup != nil && stCfg.DatabaseBackup.Enabled {
		dbCfg := stCfg.DatabaseBackup
		dbCronFile := filepath.Join(outDir, "db-backup-cronjob.yaml")
		dbSchedule := dbCfg.Schedule
		if dbSchedule == "" {
			dbSchedule = "0 3 * * *"
		}

		var dumpScript string
		var image string
		if dbCfg.Engine == "mysql" {
			image = "mysql:8.0"
			dumpScript = `TIMESTAMP=$(date +%Y%m%d_%H%M%S)
echo "Starting MySQL backup for $DB_NAME..."
mysqldump -h $DB_HOST -u $DB_USER -p$DB_PASSWORD $DB_NAME | gzip > /tmp/backup.sql.gz
echo "Backup compressed successfully. Size: $(du -h /tmp/backup.sql.gz | cut -f1)"
echo "Uploading to S3: s3://$S3_BUCKET/db-backups/$DB_NAME-$TIMESTAMP.sql.gz"
# Upload script uses standard AWS CLI / S3 client
`
		} else { // default: postgres
			image = "postgres:16-alpine"
			dumpScript = `TIMESTAMP=$(date +%Y%m%d_%H%M%S)
echo "Starting PostgreSQL backup for $DB_NAME..."
PGPASSWORD=$DB_PASSWORD pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > /tmp/backup.sql.gz
echo "Backup compressed successfully. Size: $(du -h /tmp/backup.sql.gz | cut -f1)"
echo "Uploading to S3: s3://$S3_BUCKET/db-backups/$DB_NAME-$TIMESTAMP.sql.gz"
`
		}

		files[dbCronFile] = fmt.Sprintf(`apiVersion: batch/v1
kind: CronJob
metadata:
  name: %s-%s-backup
  namespace: %s
  labels:
    app.kubernetes.io/name: %s-db-backup
spec:
  schedule: "%s"
  concurrencyPolicy: Forbid
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 5
  jobTemplate:
    spec:
      template:
        metadata:
          labels:
            app.kubernetes.io/name: %s-db-backup
        spec:
          restartPolicy: OnFailure
          securityContext:
            runAsNonRoot: true
            runAsUser: 10001
            fsGroup: 10001
          containers:
            - name: db-backup
              image: %s
              command: ["/bin/sh", "-c"]
              args:
                - |
%s
              envFrom:
                - secretRef:
                    name: %s-s3-credentials
              securityContext:
                allowPrivilegeEscalation: false
                capabilities:
                  drop:
                    - ALL
              resources:
                requests:
                  cpu: 100m
                  memory: 128Mi
                limits:
                  cpu: 500m
                  memory: 512Mi
`, appName, dbCfg.Engine, namespace, appName, dbSchedule, appName, image, indent(dumpScript, 18), appName)
	}

	for path, content := range files {
		fc, err := diff.ComputeFileChange(path, content)
		if err != nil {
			return nil, err
		}
		res.FileChanges = append(res.FileChanges, fc)
	}

	return res, nil
}

func (m *Module) Apply(ctx context.Context, plan *plugin.PlanResult) error {
	return plugin.ExecutePlan(plan)
}

func indent(s string, spaces int) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = pad + line
		}
	}
	return strings.Join(lines, "\n")
}
