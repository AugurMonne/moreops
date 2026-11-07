package storage

import (
	"context"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestStorageModuleMinIOAndBackups(t *testing.T) {
	cfg := config.NewDefaultConfig("catalog-srv")
	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled: true,
		Tool:    "helm",
	}
	cfg.Modules.Storage = &config.StorageModuleConfig{
		Enabled:  true,
		Provider: "minio-dev",
		Bucket:   "catalog-backups",
		Velero: &config.VeleroConfig{
			Enabled:  true,
			Schedule: "0 2 * * *",
			TTL:      "720h",
		},
		DatabaseBackup: &config.DBBackupConfig{
			Enabled:  true,
			Engine:   "postgres",
			Schedule: "0 4 * * *",
			Database: "catalog",
		},
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if plan == nil {
		t.Fatalf("expected non-nil plan")
	}

	foundMinIO := false
	foundVelero := false
	foundDBCron := false
	foundS3Secret := false

	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, "minio-dev.yaml") {
			foundMinIO = true
			if !strings.Contains(fc.NewData, "minioadmin") {
				t.Errorf("expected minioadmin credentials in minio-dev.yaml")
			}
		}
		if strings.Contains(fc.Path, "velero-backup.yaml") {
			foundVelero = true
			if !strings.Contains(fc.NewData, "kind: BackupStorageLocation") || !strings.Contains(fc.NewData, "kind: Schedule") {
				t.Errorf("expected BackupStorageLocation and Schedule in velero-backup.yaml")
			}
		}
		if strings.Contains(fc.Path, "db-backup-cronjob.yaml") {
			foundDBCron = true
			if !strings.Contains(fc.NewData, "kind: CronJob") || !strings.Contains(fc.NewData, "pg_dump") {
				t.Errorf("expected CronJob with pg_dump in db-backup-cronjob.yaml")
			}
		}
		if strings.Contains(fc.Path, "s3-credentials-secret.yaml") {
			foundS3Secret = true
		}
	}

	if !foundMinIO {
		t.Errorf("missing minio-dev.yaml")
	}
	if !foundVelero {
		t.Errorf("missing velero-backup.yaml")
	}
	if !foundDBCron {
		t.Errorf("missing db-backup-cronjob.yaml")
	}
	if !foundS3Secret {
		t.Errorf("missing s3-credentials-secret.yaml")
	}
}

func TestStorageModuleRussianCloudProvider(t *testing.T) {
	cfg := config.NewDefaultConfig("prod-app")
	cfg.Modules.Storage = &config.StorageModuleConfig{
		Enabled:  true,
		Provider: "yandex",
		Endpoint: "https://storage.yandexcloud.net",
		Region:   "ru-central1",
		Bucket:   "prod-app-yandex-backups",
		Velero: &config.VeleroConfig{
			Enabled:  true,
			Schedule: "0 1 * * *",
		},
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "prod")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	foundYandexBSL := false
	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, "velero-backup.yaml") {
			if strings.Contains(fc.NewData, "https://storage.yandexcloud.net") && strings.Contains(fc.NewData, "ru-central1") {
				foundYandexBSL = true
			}
		}
	}

	if !foundYandexBSL {
		t.Errorf("expected Yandex Cloud Object Storage endpoint and region in Velero config")
	}
}

func TestStorageModuleDisabled(t *testing.T) {
	cfg := config.NewDefaultConfig("app")
	cfg.Modules.Storage = nil

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != nil {
		t.Errorf("expected nil plan when storage module is disabled")
	}
}
