package vault

import "fmt"

// Config holds Vault connection parameters.
type Config struct {
	Address string
	Token   string
	Role    string
}

// SecretStoreProvider returns ESO provider block for Vault.
func SecretStoreProvider(address, authMethod string) string {
	return fmt.Sprintf(`    vault:
      server: "%s"
      path: "secret"
      version: "v2"
      auth:
        tokenSecretRef:
          name: vault-token
          key: token`, address)
}
