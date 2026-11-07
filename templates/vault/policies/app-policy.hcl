# Least-privilege policy for application secrets
path "secret/data/{{ .AppName }}/*" {
  capabilities = ["read"]
}

path "secret/data/{{ .AppName }}" {
  capabilities = ["read"]
}
