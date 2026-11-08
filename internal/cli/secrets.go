package cli

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:     "secrets",
	Aliases: []string{"secret"},
	Short:   "Manage, template and encrypt project secrets",
	Long: `Manage secrets securely without committing plaintext passwords to Git.

Supports External Secrets Operator (ESO) templates (Vault, AWS, GCP, Azure)
and file encryption via SOPS or native AES-256-GCM.`,
}

var secretsInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize local secret encryption key and SOPS configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyDir := ".moreops"
		if err := os.MkdirAll(keyDir, 0700); err != nil {
			return err
		}

		keyPath := filepath.Join(keyDir, "secrets.key")
		if _, err := os.Stat(keyPath); err == nil {
			cmd.Printf("✓ Secret encryption key already exists at %s\n", keyPath)
		} else {
			rawKey := make([]byte, 32)
			if _, err := rand.Read(rawKey); err != nil {
				return fmt.Errorf("failed generating random key: %w", err)
			}
			hexKey := hex.EncodeToString(rawKey)
			if err := os.WriteFile(keyPath, []byte(hexKey), 0600); err != nil {
				return fmt.Errorf("failed writing %s: %w", keyPath, err)
			}
			cmd.Printf("✓ Generated native AES-256 key at %s (permissions: 0600)\n", keyPath)
		}

		// Ensure key is in .gitignore
		ensureGitIgnore(".moreops/secrets.key")
		ensureGitIgnore(".moreops/*.key")

		// Create standard .sops.yaml if missing
		sopsConfig := ".sops.yaml"
		if _, err := os.Stat(sopsConfig); os.IsNotExist(err) {
			sopsData := `creation_rules:
  - path_regex: .*\.enc\.ya?ml$
    encrypted_regex: '^(data|stringData|secrets|password|token)$'
`
			_ = os.WriteFile(sopsConfig, []byte(sopsData), 0644)
			cmd.Println("✓ Created default .sops.yaml configuration")
		}

		cmd.Println("\nSecrets subsystem initialized successfully!")
		return nil
	},
}

var secretsTemplateBackend string
var secretsTemplateNamespace string

var secretsTemplateCmd = &cobra.Command{
	Use:   "template [name]",
	Short: "Generate ExternalSecret and SecretStore manifests",
	RunE: func(cmd *cobra.Command, args []string) error {
		name := "app-secrets"
		if len(args) > 0 {
			name = args[0]
		}
		ns := secretsTemplateNamespace
		if ns == "" {
			ns = "default"
		}

		manifestDir := "k8s"
		if _, err := os.Stat("charts"); err == nil {
			// Find first chart directory
			entries, _ := os.ReadDir("charts")
			for _, e := range entries {
				if e.IsDir() {
					manifestDir = filepath.Join("charts", e.Name(), "templates")
					break
				}
			}
		}

		_ = os.MkdirAll(manifestDir, 0755)
		targetFile := filepath.Join(manifestDir, name+"-externalsecret.yaml")

		var backendSpec string
		switch strings.ToLower(secretsTemplateBackend) {
		case "aws":
			backendSpec = `    aws:
      service: SecretsManager
      region: eu-central-1
      auth:
        jwt:
          serviceAccountRef:
            name: eso-sa`
		case "gcp":
			backendSpec = `    gcpsm:
      projectID: my-gcp-project
      auth:
        secretRef:
          secretAccessKeySecretRef:
            name: gcpsm-secret
            key: secret-access-credentials`
		case "azure":
			backendSpec = `    azurekv:
      authType: ManagedIdentity
      vaultUrl: "https://my-vault.vault.azure.net"`
		default: // vault
			backendSpec = `    vault:
      server: "http://vault.default.svc.cluster.local:8200"
      path: "secret"
      version: "v2"
      auth:
        kubernetes:
          mountPath: "kubernetes"
          role: "app-role"`
		}

		content := fmt.Sprintf(`# External Secrets Operator configuration
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: %s-store
  namespace: %s
spec:
  provider:
%s
---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: %s
  namespace: %s
spec:
  refreshInterval: "1h"
  secretStoreRef:
    name: %s-store
    kind: SecretStore
  target:
    name: %s-k8s
    creationPolicy: Owner
  data:
    - secretKey: DB_PASSWORD
      remoteRef:
        key: %s
        property: password
`, name, ns, backendSpec, name, ns, name, name, name)

		if err := os.WriteFile(targetFile, []byte(content), 0644); err != nil {
			return err
		}

		cmd.Printf("✓ Created ExternalSecret manifest at '%s' (provider: %s)\n", targetFile, secretsTemplateBackend)
		return nil
	},
}

var secretsEncryptCmd = &cobra.Command{
	Use:   "encrypt [file]",
	Short: "Encrypt a secrets file using SOPS or native AES-256-GCM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", filePath, err)
		}

		if _, lookErr := exec.LookPath("sops"); lookErr == nil {
			c := exec.Command("sops", "-e", "-i", filePath)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			if err := c.Run(); err == nil {
				cmd.Printf("✓ Encrypted '%s' with SOPS\n", filePath)
				return nil
			}
		}

		// Native encryption fallback
		key, err := getOrCreateNativeKey()
		if err != nil {
			return err
		}

		ciphertext, err := encryptAESGCM(data, key)
		if err != nil {
			return fmt.Errorf("native encryption failed: %w", err)
		}

		header := "# MOREOPS_ENCRYPTED_AES256_GCM\n"
		if err := os.WriteFile(filePath, []byte(header+hex.EncodeToString(ciphertext)+"\n"), 0600); err != nil {
			return err
		}

		cmd.Printf("✓ Encrypted '%s' with native AES-256-GCM\n", filePath)
		return nil
	},
}

var secretsDecryptCmd = &cobra.Command{
	Use:   "decrypt [file]",
	Short: "Decrypt a secrets file using SOPS or native AES-256-GCM",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("cannot read file %s: %w", filePath, err)
		}

		strData := string(data)
		if strings.HasPrefix(strData, "# MOREOPS_ENCRYPTED_AES256_GCM\n") {
			key, err := getOrCreateNativeKey()
			if err != nil {
				return err
			}

			payloadHex := strings.TrimSpace(strings.TrimPrefix(strData, "# MOREOPS_ENCRYPTED_AES256_GCM\n"))
			rawCiphertext, err := hex.DecodeString(payloadHex)
			if err != nil {
				return fmt.Errorf("corrupt ciphertext in %s: %w", filePath, err)
			}

			plaintext, err := decryptAESGCM(rawCiphertext, key)
			if err != nil {
				return fmt.Errorf("decryption failed (wrong key or corrupted data): %w", err)
			}

			if err := os.WriteFile(filePath, plaintext, 0600); err != nil {
				return err
			}

			cmd.Printf("✓ Decrypted '%s' using native AES-256-GCM\n", filePath)
			return nil
		}

		if _, lookErr := exec.LookPath("sops"); lookErr == nil {
			c := exec.Command("sops", "-d", "-i", filePath)
			c.Stdout = cmd.OutOrStdout()
			c.Stderr = cmd.ErrOrStderr()
			if err := c.Run(); err == nil {
				cmd.Printf("✓ Decrypted '%s' with SOPS\n", filePath)
				return nil
			}
		}

		return fmt.Errorf("file %s is not in recognized encrypted format", filePath)
	},
}

func getOrCreateNativeKey() ([]byte, error) {
	keyPath := filepath.Join(".moreops", "secrets.key")
	if data, err := os.ReadFile(keyPath); err == nil {
		hexStr := strings.TrimSpace(string(data))
		key, err := hex.DecodeString(hexStr)
		if err == nil && len(key) == 32 {
			return key, nil
		}
	}

	_ = os.MkdirAll(".moreops", 0700)
	rawKey := make([]byte, 32)
	if _, err := rand.Read(rawKey); err != nil {
		return nil, err
	}
	hexKey := hex.EncodeToString(rawKey)
	_ = os.WriteFile(keyPath, []byte(hexKey), 0600)
	ensureGitIgnore(".moreops/secrets.key")
	return rawKey, nil
}

func encryptAESGCM(plaintext []byte, key []byte) ([]byte, error) {
	h := sha256.Sum256(key)
	block, err := aes.NewCipher(h[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAESGCM(ciphertext []byte, key []byte) ([]byte, error) {
	h := sha256.Sum256(key)
	block, err := aes.NewCipher(h[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, cipherData := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, cipherData, nil)
}

func ensureGitIgnore(entry string) {
	giPath := ".gitignore"
	data, _ := os.ReadFile(giPath)
	content := string(data)
	if !strings.Contains(content, entry) {
		f, err := os.OpenFile(giPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			defer f.Close()
			if len(content) > 0 && !strings.HasSuffix(content, "\n") {
				_, _ = f.WriteString("\n")
			}
			_, _ = f.WriteString(entry + "\n")
		}
	}
}

func init() {
	secretsTemplateCmd.Flags().StringVarP(&secretsTemplateBackend, "backend", "b", "vault", "Secret store backend (vault, aws, gcp, azure)")
	secretsTemplateCmd.Flags().StringVarP(&secretsTemplateNamespace, "namespace", "n", "default", "Kubernetes namespace")

	secretsCmd.AddCommand(secretsInitCmd)
	secretsCmd.AddCommand(secretsTemplateCmd)
	secretsCmd.AddCommand(secretsEncryptCmd)
	secretsCmd.AddCommand(secretsDecryptCmd)
	rootCmd.AddCommand(secretsCmd)
}
