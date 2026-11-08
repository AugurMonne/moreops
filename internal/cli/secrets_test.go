package cli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecretsEncryptDecryptRoundTrip(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	secretFile := filepath.Join(tmpDir, "values-secret.yaml")
	originalSecret := "password: SuperSecretPassword123!\ndatabase_url: postgres://admin:pass@db:5432/mydb\n"
	if err := os.WriteFile(secretFile, []byte(originalSecret), 0600); err != nil {
		t.Fatal(err)
	}

	// Encrypt
	key, err := getOrCreateNativeKey()
	if err != nil {
		t.Fatalf("getOrCreateNativeKey failed: %v", err)
	}

	encrypted, err := encryptAESGCM([]byte(originalSecret), key)
	if err != nil {
		t.Fatalf("encryptAESGCM failed: %v", err)
	}

	// Decrypt
	decrypted, err := decryptAESGCM(encrypted, key)
	if err != nil {
		t.Fatalf("decryptAESGCM failed: %v", err)
	}

	if string(decrypted) != originalSecret {
		t.Fatalf("expected decrypted string to match original.\nGot: %s\nWant: %s", string(decrypted), originalSecret)
	}
}

func TestEnsureGitIgnore(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	ensureGitIgnore(".moreops/secrets.key")
	data, err := os.ReadFile(".gitignore")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != ".moreops/secrets.key\n" {
		t.Errorf("unexpected .gitignore content: %q", string(data))
	}

	// Run again, must not duplicate
	ensureGitIgnore(".moreops/secrets.key")
	data2, _ := os.ReadFile(".gitignore")
	if string(data2) != ".moreops/secrets.key\n" {
		t.Errorf("duplicated entry in .gitignore: %q", string(data2))
	}
}
